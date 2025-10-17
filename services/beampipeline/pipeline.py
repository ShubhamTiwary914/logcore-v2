from apache_beam.options.pipeline_options import StandardOptions
from apache_beam.options.pipeline_options import SetupOptions
from apache_beam.options.pipeline_options import PipelineOptions
from apache_beam.io.gcp.bigtableio import WriteToBigTable
from google.cloud.bigtable import row
import datetime
import apache_beam as beam
import argparse
import logging
import json
import os

class CreateRowFn(beam.DoFn):
    def process(self, message):
        rowkey = message['id'].encode('utf-8')
        direct_row = row.DirectRow(rowkey)
        for k, v in message.items():
            direct_row.set_cell(
                'cf1',
                k,
                str(v).encode('utf-8'),
                timestamp=datetime.datetime.now()
            )
        direct_row.set_cell(
            'cf2',
            'relation',
            str(message['relation']).encode('utf-8'),
            timestamp=datetime.datetime.now()
        )
        yield direct_row

def getargs():
    parser = argparse.ArgumentParser()
    parser.add_argument(
        '--topic',
        required=True,
        help=(
            'Input PubSub Topic'
            '"projects/<PROJECT>/topics/<topic>."'))
    parser.add_argument(
        '--project',
        required=True,
        help='The Bigtable project ID, this can be different than your '
        'Dataflow project')
    parser.add_argument(
        '--instance',
        required=True,
        help='The Bigtable instance ID')
    parser.add_argument(
        '--table',
        required=True,
        help='The Bigtable table ID in the instance.')
    args, pipeline_args = parser.parse_known_args()
    print(f"CLI ARGS: {json.dumps(vars(args), indent=2)}\n")
    return args, pipeline_args 
   

def main():
    # (optional - auth with credentials sa file) (recommended - ADC)
    # os.environ["GOOGLE_APPLICATION_CREDENTIALS"] = "<path-to-sa-credentials>"
    known_args, pipeline_args = getargs()
    pipeOpts = PipelineOptions(
        streaming=True,
        runner="DirectRunner",
        input_topic=known_args.topic,  
        bigtable_project=known_args.project,
        bigtable_instance=known_args.instance,
        bigtable_table=known_args.table
    )
    print(f"CLI ARGS: {known_args}")
    with beam.Pipeline(options=pipeOpts) as p:
        (
            p
            | 'Read from pubsub' >> beam.io.ReadFromPubSub(topic=known_args.topic)
            | 'To Json' >> beam.Map(lambda e: json.loads(e.decode('utf-8')))
            | 'JSON to row object' >> beam.ParDo(CreateRowFn())
            # | 'Print Data' >> beam.Map(print)
            | 'Writing row object to BigTable' >> WriteToBigTable(
                project_id=known_args.project,            
                instance_id=known_args.instance,
                table_id=known_args.table)
        )
        
main()