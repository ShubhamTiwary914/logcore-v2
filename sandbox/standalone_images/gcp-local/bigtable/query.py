import os
import argparse
from google.cloud import bigtable

os.environ["BIGTABLE_EMULATOR_HOST"] = "localhost:8086"

parser = argparse.ArgumentParser(description="List rows in a Bigtable table")
parser.add_argument("--relation", required=True, help="Table name to list rows from")
args = parser.parse_args()

client = bigtable.Client(project="gcplocal-emulator", admin=True)
instance = client.instance("local")
table = instance.table(args.relation)

rows = table.read_rows()
for row in rows:
    row_data = {cf: {col.decode(): cell[0].value.decode() 
                     for col, cell in row.cells[cf].items()} 
                for cf in row.cells}
    print(f"Row key: {row.row_key.decode()}, Data: {row_data}")
