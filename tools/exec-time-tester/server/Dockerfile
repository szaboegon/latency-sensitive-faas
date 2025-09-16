FROM python:3.10-slim

WORKDIR /app

# Copy universal server
COPY server.py /app/server.py

# Copy function code
COPY main.py /app/main.py

RUN pip install flask numpy opencv-python-headless parliament

EXPOSE 8080
CMD ["python", "server.py"]
