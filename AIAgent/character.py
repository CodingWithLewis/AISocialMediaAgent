import time
from yaspin import yaspin
from yaspin.spinners import Spinners



r = input("Please input a Character Name: ")

with yaspin() as sp:
    sp.text = f"Searching Fandom for {r}"
    time.sleep(3)
    sp.ok("✅ Found Character and Saved!")


with yaspin() as sp:
    sp.text = f"Getting Image of {r}"
    time.sleep(5)
    sp.ok("✅ Found Image and Saved!")