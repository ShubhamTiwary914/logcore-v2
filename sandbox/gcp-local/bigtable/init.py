from google.cloud import bigtable
import os

os.environ["BIGTABLE_EMULATOR_HOST"] = "localhost:8086"

client = bigtable.Client(project="gcplocal-emulator", admin=True)
instance = client.instance("local")  #default instance

# create tables
table1 = instance.table("weather_sensor")
table1.create(column_families={"cf1": None})

table2 = instance.table("machine_metrics")
table2.create(column_families={"cf1": None})

table2 = instance.table("iotsink")
table2.create(column_families={"cf1": None})
