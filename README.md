# blog_agregator
A blog aggregator project in Golang as guided by boot.dev

In order to use this, you must first have GoLang and Postgresql installed on your system. (google knows how)

Install gator by running ```go install github.com/pressly/goose/v3/cmd/goose@latest``` in your command line, and follow instructions [here](http://github.com/gshalay/gator) for setting up your user and local server.

Run go build in your command prompt while in the main directory of the project, and then run ```./blog_aggregator [command]``` to run the command

After running the program once, make sure there is a config file generated and placed in your home directory called ```.gatorconfig.json```

Here is a list of runnable commands:
* ```reset```: empty all tables and known users
* ```register [username]```: register a new user
* ```login [username]```: login to an existing user
* ```users```: lists all users
* ```agg [time_between_fetches]```: begin the loop to fetch feeds
* ```addfeed [feed_name] [url]```: add and follow a new feed
* ```follow [url]```: follow an existing feed
* ```unfollow [url]```: unfollow an existing feed
* ```following```: list feeds the current user is following
* ```feeds```: list all feeds
* ```browse [opt: limit]```: show most recent posts for your feeds