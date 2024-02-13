import os
import uuid
from concurrent.futures import ThreadPoolExecutor, as_completed
from time import sleep

import psycopg2
from openai import OpenAI
import requests
from requests import RequestException, HTTPError
from urllib.parse import urlencode
import json

from tqdm import tqdm
from dotenv import load_dotenv

load_dotenv()


# Connect Proxy
http_proxy = "http://10.10.1.10:3128"
https_proxy = "https://10.10.1.11:1080"
ftp_proxy = "ftp://10.10.1.10:3128"

client = OpenAI()
model = "gpt-4-turbo-preview"
messages = {
    "role": "system",
    "content": "You are a bot that is giving text to certain keys in a dictionary. Follow instructions"
    "carefully.",
}

# character_information = []


def generate_ai_text(prompt):
    generate = client.chat.completions.create(
        model=model,
        messages=[
            messages,
            {
                "role": "user",
                "content": f"{prompt}",
            },
        ],
    )

    return generate.choices[0].message.content


def get_character_information(character_name):
    proxies = {
        "http": "http://brd-customer-hl_4a5981ec-zone-serp_api1:q6ghf9gqwiyp@brd.superproxy.io:22225",
        "https": "http://brd-customer-hl_4a5981ec-zone-serp_api1:q6ghf9gqwiyp@brd.superproxy.io:22225",
    }

    query_params = {
        "q": f"{character_name} jpg",
        "tbm": "isch",
        "brd_json": "1",
    }

    url = urlencode(query_params)
    attempts = 0
    max_attempts = 5
    backoff_factor = 1  # Starting backoff delay in seconds

    while attempts < max_attempts:
        try:
            r = requests.get(f"http://www.google.com/search?{url}", proxies=proxies)
            r.raise_for_status()

            # Check if the response contains images
            if "images" not in r.json():
                raise RequestException("No images found")

            # If this point is reached, the request was successful, and images were found
            break

        except (RequestException, HTTPError) as e:
            attempts += 1
            print(f"Attempt {attempts} failed: {e}")

            if attempts == max_attempts:
                raise Exception("Max retries reached, unable to fetch images.")

            sleep_time = backoff_factor * (2 ** (attempts - 1))  # Exponential backoff
            print(f"Retrying in {sleep_time} seconds...")
            sleep(sleep_time)

    images = r.json()["images"]

    image_base64 = images[0]["image_base64"]

    bio = generate_ai_text(f"Create a 120 character bio for {character_name}")

    location = generate_ai_text(f"Create a location for {character_name}")

    username = generate_ai_text(f"Create a username for {character_name}")

    return {
        "id": str(uuid.uuid4()),
        "bio": bio,
        "name": character_name,
        "theme": "blue",
        "accent": "blue",
        "website": None,
        "location": location,
        "username": username,
        "photoBase64": image_base64,
        "verified": True,
        "following": [],
        "followers": [],
        "totalTweets": 0,
        "totalPhotos": 0,
    }


# Save characters to neon database
conn = psycopg2.connect(
    os.getenv("DATABASE_URL"),
)

cur = conn.cursor()

insert_sql = """
INSERT INTO characters (id, bio, name, theme, accent, website, location, username, photoBase64, verified, totalTweets, totalPhotos)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
"""

with open("50characters.json", "r") as file:
    characters = json.loads(file.read())

    for character in tqdm(characters):
        cur.execute(
            insert_sql,
            (
                character["id"],
                character["bio"],
                character["name"],
                character["theme"],
                character["accent"],
                character["website"],
                "",
                character["username"],
                character["photoBase64"],
                character["verified"],
                character["totalTweets"],
                character["totalPhotos"],
            ),
        )

conn.commit()
cur.close()
conn.close()
# with open("50characters.txt", "r") as file:
#     characters = [line.strip() for line in file]
#     # Process characters in parallel
#     with ThreadPoolExecutor(max_workers=5) as executor:
#         # Submit all characters to the executor
#         future_to_character = {
#             executor.submit(get_character_information, character): character
#             for character in characters
#         }
#
#         for future in as_completed(future_to_character):
#             character = future_to_character[future]
#             try:
#                 info = future.result()
#                 character_information.append(info)
#             except Exception as exc:
#                 print(f"{character} generated an exception: {exc}")
#
#
# # Dump to json file
#
# with open("50characters.json", "w") as file:
#     json.dump(character_information, file)

# with open("50characters.json", "r") as file:
#     characters = json.load(file)
#     for character in characters:
#         r = requests.post(
#             "http://localhost:3000/user/",
#             headers={"Content-Type": "application/json"},
#             json=character,
#         )
