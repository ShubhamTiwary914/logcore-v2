import os
import sys
import yaml
import json
import logging
import argparse
import apache_beam as beam
from datetime import datetime
from google.cloud.bigtable.row import DirectRow
from apache_beam.io.gcp.pubsub import ReadFromPubSub
from apache_beam.io.gcp.bigtableio import WriteToBigTable
from apache_beam.options.pipeline_options import PipelineOptions

logger = logging.getLogger()
logger.setLevel(logging.INFO)
log_path = os.path.join(os.path.dirname(__file__), "pipeline.log")
log_file = open(log_path, "a", buffering=1)
sys.stdout = log_file  

def get_args():
    parser = argparse.ArgumentParser()
    parser.add_argument("--project", required=True)
    parser.add_argument("--topic", required=True)
    parser.add_argument("--host", required=True)
    return parser.parse_args()

def getNowTime():
    return datetime.now().strftime("%b %d %H:%M:%S")

cliargs : argparse.Namespace = get_args()
project_id : str= cliargs.project
topic : str= cliargs.topic
host : str = cliargs.host
os.environ["PUBSUB_EMULATOR_HOST"] = host
os.environ["BIGTABLE_EMULATOR_HOST"] = "localhost:8086"
os.environ["GOOGLE_CLOUD_PROJECT"] = project_id
os.environ["GOOGLE_APPLICATION_CREDENTIALS"] = "/dev/null"
instance_id = "local"

print(f"CLI_ARGS: \n{yaml.dump(vars(cliargs))}")
startTime = getNowTime()
print(f"Started at: {startTime}")

options : PipelineOptions = PipelineOptions(
    streaming=True,
    runner="DirectRunner"
)

class ToBigTableRowDynamic(beam.DoFn):
    def process(self, element):
        data = json.loads(element.data.decode("utf-8"))
        row_key = data["id"].encode("utf-8")
        #all messages will have attribute "relation" -> defines bigtable table
        table_id = data["relation"]
        row = DirectRow(row_key=row_key)
        print(f"Writing to BigTable relation:{table_id}  key:{data["id"]}...")
        for k, v in data.items():
            if k != "id":
                row.set_cell("cf1", k, str(v).encode("utf-8"))
        yield (table_id, row)


class WriteTableFn(beam.DoFn):
    def process(self, table_rows):
        table_id, rows_list = table_rows
        yield rows_list | f"Write_{table_id}" >> WriteToBigTable(
            project_id=project_id,
            instance_id=instance_id,
            table_id=table_id
        )

def main():
    try:
        with beam.Pipeline(options=options) as p:
            (
                p
                | "ReadFromPubSub" >> ReadFromPubSub(topic=topic, with_attributes=True)
                | "WindowIntoFixed" >> beam.WindowInto(beam.window.FixedWindows(3))  #streaming-window:3s
                | "ToBigTableRow" >> beam.ParDo(ToBigTableRowDynamic())
                | "GroupByTable" >> beam.GroupByKey()
                | "WriteAllTables" >> beam.ParDo(WriteTableFn())
            )
    except KeyboardInterrupt:
        endTime = getNowTime()
        print(f"\nStarted at: {startTime}\nEnded at: {endTime}\nPipeline Complete!", file=sys.__stdout__,flush=True)
        print(f"Ended at: {endTime}\n\n")
    finally:
        log_file.close()

if __name__ == '__main__':
    main()