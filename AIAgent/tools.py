from time import sleep
from typing import Union
from urllib.parse import urlencode

import requests
from langchain.pydantic_v1 import BaseModel, Field
from langchain.tools import BaseTool, StructuredTool, tool
from newspaper import Article
from requests import RequestException, HTTPError

headers = requests.utils.default_headers()


class TweetInput(BaseModel):
    tweet_content: str = Field(
        description="The content of the tweet going out into the world."
    )
    is_reply: bool = Field(
        description="Whether or not this is a reply to another tweet."
    )
    user_id: str = Field(description="The id of the user who is sending the tweet.")
    tweet_id: Union[str, None] = Field(
        description="The id of the tweet being replied to if the tweet is a reply."
    )


class ImageTweetInput(BaseModel):
    image_prompt: str = Field(
        description="The prompt for an image for an AI to generate using Stable Diffusion"
        "Please include a really detailed prompt for the AI to generate an image."
        "Make sure to use key details in style to match where you are from."
    )
    tweet_content: str = Field(
        description="The content of the tweet going out into the world."
    )
    is_reply: bool = Field(
        description="Whether or not if this is a reply to an existing tweet or not."
    )
    user_id: str = Field(description="The id of the user who is sending the tweet.")
    tweet_id: Union[str, None] = Field(
        description="The id of the tweet being replied to if the tweet is a reply."
    )


@tool("tweet-tool", args_schema=TweetInput)
def tweet(tweet_content: str, is_reply: bool, user_id: str, tweet_id=None) -> str:
    """Send a tweet online!"""

    reply_to = None
    if is_reply:
        reply_to = {"id": tweet_id, "username": "Testing"}

    r = requests.post(
        "http://localhost:3000/tweet/",
        headers=headers,
        json={
            "text": tweet_content,
            "parent": reply_to,
            "createdBy": user_id,
            "createdAt": "2021-01-01T00:00:00.000Z",
            "userReplies": 0,
            "userRetweets": [],
            "userLikes": [],
        },
    )

    return r.text


class NewsInput(BaseModel):
    news_item: str = Field(
        description="A hobby / interest / topic that you want to find the latest news about. This will be used to find"
        "an article that is returned."
    )


@tool("get-latest-news-to-tweet-about", args_schema=NewsInput)
def get_latest_news_to_tweet_about(news_item):
    """Get the latest news to tweet about."""
    proxies = {
        "http": "http://brd-customer-hl_4a5981ec-zone-serp_api1:q6ghf9gqwiyp@brd.superproxy.io:22225",
        "https": "http://brd-customer-hl_4a5981ec-zone-serp_api1:q6ghf9gqwiyp@brd.superproxy.io:22225",
    }

    query_params = {
        "q": f"{news_item} news",
        "tbm": "nws",
        "brd_json": "1",
    }

    url = urlencode(query_params)
    attempts = 0
    max_attempts = 5
    backoff_factor = 1  # Starting backoff delay in seconds

    while attempts < max_attempts:
        try:
            r = requests.get(
                f"http://www.google.com/search?{url}", headers=headers, proxies=proxies
            )
            r.raise_for_status()

            # Check if the response contains images
            if "news" not in r.json():
                raise RequestException("No news found")

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

    news = r.json()["news"][0]["link"]

    article = Article(news)
    article.download()
    article.parse()
    return {"content": article.text, "link": news}


@tool("get-latest-tweets")
def get_latest_tweets() -> str:
    """Get the latest tweets from the server."""
    return requests.get("http://localhost:3000/tweets/", headers=headers).text


@tool("generate-image-tweet", args_schema=ImageTweetInput)
def generate_image_tweet(
    image_prompt: str, tweet_content: str, is_reply: bool, user_id: str, tweet_id=None
) -> str:
    """Create a prompt that will generate an image using AI to go alongside a tweet."""

    reply_to = None
    if is_reply:
        reply_to = {"id": tweet_id, "username": "Testing"}

    r = requests.post(
        "http://localhost:3000/image-tweet/",
        headers=headers,
        json={
            "imagePrompt": image_prompt,
            "tweet": {
                "text": tweet_content,
                "parent": reply_to,
                "createdBy": user_id,
                "createdAt": "2021-01-01T00:00:00.000Z",
                "userReplies": 0,
                "userRetweets": [],
                "userLikes": [],
            },
        },
    )

    if not r.ok:
        return f"Error: {r.status_code} - {r.text}"

    return "Image and tweet created successfully!"
