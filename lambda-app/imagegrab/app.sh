#!/bin/sh

exec opentelemetry-instrument python -m parliament "$(dirname "$0")"
