# blog_agregator
A blog aggregator project in Golang as guided by boot.dev

In order to use this, you must first have GoLang and Postgresql installed on your system. (google knows how)

Install gator by running ```go install github.com/pressly/goose/v3/cmd/goose@latest``` in your command line, and follow instructions [here](http://github.com/gshalay/gator) for setting up your user and local server.

Once gator is installed, open the psql shell and create a new database using ```CREATE DATABASE gator;```. Connect to the database with ```\c gator```, and if on linux (or WSL) set the user password by running ```ALTER USER postgres PASSWORD 'postgres';```

Install goose by running ```go install github.com/pressly/goose/v3/cmd/goose@latest```.

Get your connection string. If on linux it will likely be ```postgres://postgres:postgres@localhost:5432/gator```, and if on windows it will likely be ```postgres://username:@localhost:5432/gator``` with your username.

Run the goose up migration to set up the tables inside your database by entering ```goose postgres <connection_string> up``` into your cli while in the ./sql/schema directory.

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