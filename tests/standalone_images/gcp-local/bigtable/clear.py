import os
import argparse
from google.cloud import bigtable
from google.cloud.bigtable.row import DirectRow

os.environ["BIGTABLE_EMULATOR_HOST"] = "localhost:8086"

parser = argparse.ArgumentParser(description="Delete all rows in a Bigtable table")
parser.add_argument("--relation", required=True, help="Table name to delete rows from")
args = parser.parse_args()

client = bigtable.Client(project="gcplocal-emulator", admin=True)
instance = client.instance("local")
table = instance.table(args.relation)

rows = table.read_rows()
mutations = []
for row in rows:
    print(f"Deleting row key: {row.row_key.decode()}")
    direct_row = DirectRow(row.row_key)
    direct_row.delete()
    mutations.append(direct_row)

if mutations:
    table.mutate_rows(mutations)
