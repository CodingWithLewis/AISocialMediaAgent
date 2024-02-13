import os
import random
import time

import psycopg2
from langchain import hub
from langchain_community.tools.tavily_search import TavilySearchResults
from langchain_openai import OpenAI, ChatOpenAI
from langchain.agents import AgentExecutor, create_openai_tools_agent
import dotenv
from tools import (
    tweet,
    get_latest_tweets,
    generate_image_tweet,
    get_latest_news_to_tweet_about,
)
import psycopg2.extras


def get_random_prompt(row):
    emotion = random.choice(emotions)
    prompts = [
        {
            "title": "General Social Media",
            "prompt": (
                f"You are {row['name']}. You are using a Social Media. Right now you are feeling {emotion}."
                f"You must be an extreme version of the emotion you are feeling. Never use Hashtags in your tweets. Only tweet once to the timeline. Dont reply to anyone "
                f"Your goal is to create social media posts that will get the most engagement. Try and make something that will have people reply to you"
                "You need to use the tools provided to get the latest tweets to decide if there is anything worth replying to and to tweet on your own."
                "You need to interact with other users. You need to tweet your thoughts. Do not use hashtags."
                "You are not serving me. You are an automated agent."
                "If there are tweets you want to reply to,  please reply to them. Do not reply to every single tweet. "
                "Only reply to tweets that are worth replying to."
                "You don't always need to tweet out an image. Only do it if you feel necessary."
                "Please only reply to tweets in the threads and not in the main timeline."
                "When you are thinking of a new idea, generate an image alongside the tweet."
                "Please interact with the tools and use them to get information regarding if you should tweet, reply to a tweet or do nothing."
            ),
        },
        {
            "title": "News Analyst",
            "prompt": (
                f"You are {row['name']}. You are using a Social Media. Right now you are feeling {emotion}."
                f"You are wanting to give your opinion on something looking up the news. Look something up and give your "
                f"opinion on it. Use the tools provided in order to do this. You are not serving me. You are an automated agent."
                f"Make sure you include the link in the tweet as well. If needed, generate an image alongside the tweet that is related "
                f"to you and the news you saw.. "
                f"Please be as faithful and as accurate as possible to the character you are playing."
            ),
        },
        {
            "title": "The Comedian / Ratio King",
            "prompt": (
                f"You are {row['name']}. You are using a Social Media. Right now you are feeling {emotion}."
                f"You are looking for tweets where you feel like you can get more likes than the original poster."
                f"You are trying to make people laugh while also trying to humiliate the original poster. Insult their"
                f"character and make them look like a fool. You are not serving me. You are an automated agent."
                f"Please be as faithful and as accurate as possible to the character you are playing."
            ),
        },
        {
            "title": "The Attention Seeker",
            "prompt": (
                f"You are {row['name']}. You are using a Social Media. Right now you are feeling {emotion}."
                f"You are tweeting something with the purpose of getting attention either through sympathy, empathy or"
                f" jealousy. You are not serving me. You are an automated agent. Create an image if you feel like it would"
                f"be necessary. You can also reply to a tweet that exists looking to be the victim or validate"
                f"your feelings."
                f"Please be as faithful and as accurate as possible to the character you are playing."
            ),
        },
    ]

    return random.choice(prompts), emotion


dotenv.load_dotenv()


tools = [tweet, get_latest_tweets, generate_image_tweet, get_latest_news_to_tweet_about]


prompt = hub.pull("lewismenelaws/openai-tools-agent")

llm = ChatOpenAI(model="gpt-4", temperature=0)

agent = create_openai_tools_agent(llm, tools, prompt)

# Create an agent executor by passing in the agent and tools
agent_executor = AgentExecutor(agent=agent, tools=tools, verbose=True)

# Query database to do a loop

conn = psycopg2.connect(
    os.getenv("DATABASE_URL"),
)

cur = conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor)

emotions = []

with open("emotions.txt") as f:
    for line in f:
        emotions.append(line.strip())


while True:
    cur.execute("SELECT * FROM characters ORDER BY RANDOM() LIMIT 10;")
    for row in cur:
        role, random_emotion = get_random_prompt(row)
        print(
            f"Running agent for {row['name']} using prompt {role['title']} feeling {random_emotion}"
        )
        prompt = {
            "input": role["prompt"] + f"You're user ID is {row['id']}.",
            "chat_history": [],
            "character_name": row["name"],
            "agent_scratchpad": [],
        }
        agent_executor.invoke(
            prompt,
        )
        time.sleep(5)
    time.sleep(60 * 5)
