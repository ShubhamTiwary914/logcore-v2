import os
import yaml
import requests
import argparse
import json


# ----------------------------
# region CLI parsing
# ----------------------------
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



# ----------------------------
# region Environment variables
# ----------------------------
os.environ["PUBSUB_EMULATOR_HOST"] = cliargs.host
os.environ["GOOGLE_CLOUD_PROJECT"] = cliargs.project
os.environ["BIGQUERY_EMULATOR_HOST"] = "http://localhost:9050"
os.environ["GOOGLE_APPLICATION_CREDENTIALS"] = "/home/dev/work/projects/logcore-gcp/tests/standalone_images/gcp-local/directrunner/fakecreds.json"
BQHOST=os.getenv('BIGQUERY_EMULATOR_HOST')


# ----------------------------
# region Beam Pipeline (after patch)
# ----------------------------
import apache_beam as beam
from apache_beam.options.pipeline_options import PipelineOptions
from apache_beam.io.gcp.pubsub import ReadFromPubSub
from apache_beam.io.gcp.bigquery import WriteToBigQuery
from apache_beam.io.gcp.bigquery_tools import BigQueryWrapper

options : PipelineOptions = PipelineOptions(
    streaming=True,
    runner="DirectRunner"
)

def convertBQschema(dict_schemas: dict) -> dict:
    """
    Convert {table: [ {name, type}, ... ]} into
    {table: "name:TYPE,name:TYPE,..."} for Beam BigQuery sink.
    """
    converted = {}
    for table, fields in dict_schemas.items():
        converted[table] = ",".join(f"{f['name']}:{f['type']}" for f in fields)
    return converted

# ----------------------------
# region Helper Methods
# ----------------------------
def getSchemasRaw() -> dict:
    with open("schemas.json", "r") as f:
        return json.load(f)
SCHEMAS = convertBQschema(getSchemasRaw())


def list_topics_pubsub(project) -> list[str]:
    url = f"http://{os.environ['PUBSUB_EMULATOR_HOST']}/v1/projects/{project}/topics"
    topics = []
    while url:
        r = requests.get(url)
        data = r.json()
        topics.extconvert_schemasend([t['name'] for t in data.get('topics', [])])
        next_token = data.get('nextPageToken')
        if next_token:
            url = f"http://{os.environ['PUBSUB_EMULATOR_HOST']}/v1/projects/{project}/topics?pageToken={next_token}"
        else:
            url = None
    return topic


# ----------------------------
# region Pipeline & DoFns
# ----------------------------
class GetRelationFn(beam.DoFn):
    def process(self, element):
        relation = element.attributes.get('relation') if hasattr(element, 'attributes') else None
        print(f"Relation: {relation}\n")

class GetDataFn(beam.DoFn):
    def process(self, element):
        data = json.loads(element.data.decode('utf-8'))
        yield data


def main():
    with beam.Pipeline(options=options) as p:
        (
            p
            | "ReadFromPubSub" >> ReadFromPubSub(topic=topic, with_attributes=True)
            | "WindowIntoFixed" >> beam.WindowInto(beam.window.FixedWindows(3))  #streaming-window:3s
        )
        


if __name__ == '__main__':
    main()