from google.cloud import bigtable
import os

os.environ["BIGTABLE_EMULATOR_HOST"] = "localhost:8086"

client = bigtable.Client(project="gcplocal-emulator", admin=True)
instance = client.instance("local")

tables = instance.list_tables()
for table in tables:
    print(table.table_id)
