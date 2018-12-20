# Analysis of Hacker News Who is Hiring

This is the very initial version of the code. The goal of the project was to generate analytics and visualize them on the charts, so there are no tests at the moment. Feel free to fork or send a PR if logic for any of the analytics doesn't look good.

#### Code

The `main.go` does the following,

* Scrape all comments from Hacker News "Who is Hiring" threads. I have hard coded a map which contains the month to monthly thread id(item id).
* Generates counts for Remote work, Titles, Visa, Lunch Perks
* Generates skill counts.

The `places.py` uses `GeoCities` (you need t0 install it before it works) module to generate locations from the text, aggregates the counts and stores the value in a json file.

#### Charts

The charts are all created from the data from the above code. Below are some code pen links which has the code for the charts,

https://codepen.io/vthallam/pen/wRoORO - Map Chart with Bubble Data Points, made using amCharts.
https://codepen.io/vthallam/pen/ZVLzvq - Bar chart made using Highchart 



