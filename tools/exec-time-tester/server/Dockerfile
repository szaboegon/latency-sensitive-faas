FROM python:3.10-slim

WORKDIR /app

# Copy universal server
COPY server.py /app/server.py

# Copy function code
COPY main.py /app/main.py

# Install dependencies
COPY requirements.txt /app/requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

EXPOSE 8080
CMD ["python", "server.py"]
