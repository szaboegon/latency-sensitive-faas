@echo off

set OTELCOLLECTOR_HELMCHART_VALUES="helmcharts\otelcollector.values.yaml"
set ES_HELMCHART_VALUES="helmcharts\elasticsearch.values.yaml"
set KIBANA_HELMCHART_VALUES="helmcharts\kibana.values.yaml"
set APM_SERVER_HELMCHART_VALUES="helmcharts\apm-server.values.yaml"
set METRICBEAT_HELMCHART_VALUES="helmcharts\metricbeat.values.yaml"

echo Creating observability namespace...
kubectl create namespace observability
if errorlevel 1 (
    echo Error creating namespace observability
    exit /b 1
)

@REM helm install otel-collector open-telemetry/opentelemetry-collector -n observability --values %OTELCOLLECTOR_HELMCHART_VALUES%
@REM if errorlevel 1 (
@REM     echo Error installing otel collector from helm chart
@REM     exit /b 1
@REM )

helm install elasticsearch elastic/elasticsearch -n observability --values %ES_HELMCHART_VALUES%
if errorlevel 1 (
    echo Error installing elasticsearch
    exit /b 1
)

helm install kibana elastic/kibana -n observability --values %KIBANA_HELMCHART_VALUES%
if errorlevel 1 (
    echo Error install kibana
    exit /b 1
)

helm install apm-server elastic/apm-server -n observability --values %APM_SERVER_HELMCHART_VALUES%
if errorlevel 1 (
    echo Error installing apm-server
    exit /b 1
)

helm install metricbeat elastic/metricbeat -n observability --values %METRICBEAT_HELMCHART_VALUES%
if errorlevel 1 (
    echo Error installing metricbeat
    exit /b 1
)

