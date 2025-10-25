import base64
import json
import time
import traceback
import boto3
from backend.feature_extraction.extractor import FeatureExtractor
from backend.version import WATERMARK

# 🔧 Config
DEBUG = True
DELIVERY_STREAM = "ai-steward-ingestion-stream"
REGION = "us-east-1"

# 🔌 Clients
cloudwatch = boto3.client("cloudwatch", region_name=REGION)
firehose = boto3.client("firehose", region_name=REGION)

# 📊 Metric Publisher
def publish_metric(name, value, unit="Count"):
    try:
        cloudwatch.put_metric_data(
            Namespace="ai-steward-ingestion",
            MetricData=[{
                "MetricName": name,
                "Value": value,
                "Unit": unit
            }]
        )
        if DEBUG:
            print(f"[Metric] {name} = {value} {unit}")
    except Exception as metric_err:
        print(f"[MetricError] Failed to publish {name}: {metric_err}")

# 🚀 Lambda Entry Point
def lambda_handler(event, context):
    start_time = time.time()
    print("🔥 Lambda invoked")

    try:
        if DEBUG:
            print(f"[Event] {json.dumps(event)}")

        # 🧬 Decode base64 NDJSON payload
        encoded_data = event.get("Record", {}).get("Data")
        if not encoded_data:
            raise ValueError("Missing 'Record.Data' in event")

        decoded = base64.b64decode(encoded_data).decode("utf-8")
        telemetry = json.loads(decoded)

        if DEBUG:
            print("[Decode] Telemetry parsed")
            print(json.dumps(telemetry, indent=2))

        # 🧠 Feature Extraction
        extractor = FeatureExtractor(telemetry, watermark=WATERMARK)
        features = extractor.extract()

        if DEBUG:
            print("[Features] Extracted")
            print(json.dumps(features, indent=2))

        # 🏗️ Build structured incident
        incident = {
            "incident_id": telemetry.get("incident_id"),
            "car_id": telemetry.get("car_id"),
            "impact_zone": telemetry.get("impact_zone"),
            "features": features,
            "timestamp": telemetry.get("timestamp")
        }

        # 🚚 Deliver to Firehose
        try:
            response = firehose.put_record(
                DeliveryStreamName=DELIVERY_STREAM,
                Record={"Data": json.dumps(incident) + "\n"}
            )
            if DEBUG:
                print("[Firehose] Record delivered")
                print("[Firehose] Response:", json.dumps(response, indent=2))
        except Exception as fh_err:
            print(f"[FirehoseError] {str(fh_err)}")
            raise fh_err

        # 📊 Metrics
        publish_metric("IngestionSuccess", 1)
        publish_metric("IngestionLatency", time.time() - start_time, unit="Seconds")

        print("[Incident] Structured:")
        print(json.dumps(incident, indent=2))

        return {"status": "success", "incident": incident}

    except Exception as e:
        publish_metric("IngestionFailure", 1)
        publish_metric("IngestionLatency", time.time() - start_time, unit="Seconds")

        error_trace = traceback.format_exc()
        print(f"[ERROR] {str(e)}")
        print(f"[TRACE] {error_trace}")

        return {"error": str(e), "trace": error_trace}