FROM python:3.10.13-bullseye

WORKDIR /app

COPY . /app

RUN apt-get update && apt-get install -y python3-pip

RUN pip3 install --upgrade pip

RUN pip3 install --upgrade setuptools

RUN pip3 install -r requirements.txt

CMD ["python3", "main.py"]