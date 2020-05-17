# scavenge

This is a Telegram bot to support my little lockdown scavenger hunt game for
friends and family.

## Chat API

```
Bot: Hello, I am the Lockdown Scavenger Hunt bot. Type "help" to find out how to play.

Human: help

Bot: The Lockdown Scavenger Hunt is a timed event where you will be given a
     list of items to find and photograph within an hour. All members of your team
     can send me photographs and I'll build them into an album for judging once
     the hour is over. The Lockdown Scavenger Hunt is just for fun, and any
     rumours of prizes are just that, rumours...

     I understand the following commands:
     - /createteam: create a new team (you'll need to tell me a team name)
     - /jointeam: join an existing team (you'll need to tell me that same team name)
     - /leaveteam: leave your current team (everyone makes mistakes)
     - /listitems: I'll show you the list of items to be find (but only when the game has started)
     - /time: I'll tell you how long is left in the game to submit photos

    Once the game has started you'll be able to send my photos here and I'll
    add them to your team's gallery

Human: create

Bot: What is your team name?

Human: Limekilns losers

Bot: Limekilns losers it is

Human: list

Bot: I'm afraid the game hasn't started yet, so I can't tell you the list yet




Human2: join

Bot: What team would you like to join?

Human2: Limekilns winners

Bot: I'm afraid I can't find a team with that name, please check and tell me again

Human2: Limekilns losers

Bot: Congratulations you are now part of the Limekilns losers, good luck!
```
