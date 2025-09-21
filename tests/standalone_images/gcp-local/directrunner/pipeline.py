import apache_beam as beam
from apache_beam.options.pipeline_options import PipelineOptions
from apache_beam.io.gcp.pubsub import ReadFromPubSub
import os
import yaml
import requests
import argparse
import json

def get_args():
    parser = argparse.ArgumentParser()
    parser.add_argument("--project", required=True)
    parser.add_argument("--topic", required=True)
    parser.add_argument("--host", required=True)
    return parser.parse_args()

cliargs : argparse.Namespace = get_args()
project_id : str= cliargs.project
topic : str= cliargs.topic
host: str =cliargs.host


print(f"CLI_ARGS: \n{yaml.dump(vars(cliargs))}")

os.environ["PUBSUB_EMULATOR_HOST"] = cliargs.host
os.environ["GOOGLE_CLOUD_PROJECT"] = cliargs.project

options : PipelineOptions = PipelineOptions(
    streaming=True,
    runner="DirectRunner"
)

def getSchemas() -> str:
    with open("schemas.yaml", "r") as f:
        schema_file = yaml.safe_load(f)
    schemas : str = schema_file.get("schemas", {})
    return schemas


def list_topics_pubsub(project) -> list[str]:
    url = f"http://{os.environ['PUBSUB_EMULATOR_HOST']}/v1/projects/{project}/topics"
    topics = []
    while url:
        r = requests.get(url)
        data = r.json()
        topics.extend([t['name'] for t in data.get('topics', [])])
        next_token = data.get('nextPageToken')
        if next_token:
            url = f"http://{os.environ['PUBSUB_EMULATOR_HOST']}/v1/projects/{project}/topics?pageToken={next_token}"
        else:
            url = None
    return topic


class GetRelationFn(beam.DoFn):
    def process(self, element):
        relation = element.attributes.get('relation') if hasattr(element, 'attributes') else None
        print(f"Relation: {relation}\n")

class GetDataFn(beam.DoFn):
    def process(self, element):
        data = json.loads(element.data.decode('utf-8'))
        yield data


def main():
    print(json.dumps(getSchemas(), indent=2))
    with beam.Pipeline(options=options) as p:
        stream = (
            p
            | "ReadFromPubSub" >> ReadFromPubSub(topic=topic, with_attributes=True)
            | "WindowIntoFixed" >> beam.WindowInto(beam.window.FixedWindows(3))  #streaming-window:3s
        )
        relation = stream | "GetRelation" >> beam.ParDo(GetRelationFn())
        data = stream | "GetData" >> beam.ParDo(GetDataFn())

        relation | "PrintRelations" >> beam.Map(print)


if __name__ == '__main__':
    main()