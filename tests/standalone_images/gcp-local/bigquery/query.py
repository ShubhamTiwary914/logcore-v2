from google.cloud import bigquery

client = bigquery.Client(
    project="gcplocal-emulator",
    client_options={"api_endpoint": "http://0.0.0.0:9050"},
    credentials=None  # disables auth
)

query_job = client.query("""
    SELECT 
        id, 
        CAST(timestamp AS STRING) AS timestamp_str, 
        temperature_c,
        pressure_hpa
    FROM iot_data.weather_sensor
""")
for row in query_job:
    print(row[::])
