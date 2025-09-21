import os
import yaml
import requests
import argparse
import json
import apache_beam as beam
from apache_beam.options.pipeline_options import PipelineOptions
from apache_beam.io.gcp.pubsub import ReadFromPubSub
from apache_beam.io.gcp.bigtableio import WriteToBigTable
from google.cloud.bigtable.row import DirectRow

import logging
logger = logging.getLogger()
logger.setLevel(logging.INFO)


# ----------------------------
# region CLI parsing
def get_args():
    parser = argparse.ArgumentParser()
    parser.add_argument("--project", required=True)
    parser.add_argument("--topic", required=True)
    parser.add_argument("--host", required=True)
    return parser.parse_args()

cliargs : argparse.Namespace = get_args()
project_id : str= cliargs.project
topic : str= cliargs.topic
host : str = cliargs.host
print(f"CLI_ARGS: \n{yaml.dump(vars(cliargs))}")


# ----------------------------
# region Environment variables
os.environ["PUBSUB_EMULATOR_HOST"] = host
os.environ["BIGTABLE_EMULATOR_HOST"] = "localhost:8086"
os.environ["GOOGLE_CLOUD_PROJECT"] = project_id
os.environ["GOOGLE_APPLICATION_CREDENTIALS"] = "/dev/null"
instance_id = "local"


# ----------------------------
# region Beam Pipeline (after patch)
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
def getSchemasRaw() -> dict:
    with open("../schemas.json", "r") as f:
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
class ToBigTableRowDynamic(beam.DoFn):
    def process(self, element):
        data = json.loads(element.data.decode("utf-8"))
        row_key = data["id"].encode("utf-8")
        table_id = element.attributes.get("relation")
        row = DirectRow(row_key=row_key)
        print(f"Writing to BigTable relation - {table_id} key:{row_key}...")
        for k, v in data.items():
            if k != "id":
                row.set_cell("cf1", k, str(v).encode("utf-8"))
        yield (table_id, row)


class WriteTableFn(beam.DoFn):
    def process(self, table_rows):
        table_id, rows_list = table_rows
        print(f"Writing to relation - {table_id}")
        yield rows_list | f"Write_{table_id}" >> WriteToBigTable(
            project_id=project_id,
            instance_id=instance_id,
            table_id=table_id
        )

def main():
    with beam.Pipeline(options=options) as p:
        stream = (
            p
            | "ReadFromPubSub" >> ReadFromPubSub(topic=topic, with_attributes=True)
            | "WindowIntoFixed" >> beam.WindowInto(beam.window.FixedWindows(3))  #streaming-window:3s
            | "ToBigTableRow" >> beam.ParDo(ToBigTableRowDynamic())
            | "GroupByTable" >> beam.GroupByKey()
            | "WriteAllTables" >> beam.ParDo(WriteTableFn())
        )


if __name__ == '__main__':
    main()