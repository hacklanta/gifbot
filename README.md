# GifBot

This is a simple Slack bot for storing GIFs under keywords for use later. Once this bot is up and
running, you can store a GIF under a keyword by invoking the .storegif command:

```
.gifstore nailed-it https://cdn-images-1.medium.com/max/1600/1*mvD_4BDe6H3Duz4RsmZsbg.gif
```

Then you can recall any nailed-it gif at random by just dropping the following message into
any channel that Gifbot is in:

```
.gif nailed-it
```

If multiple GIFs are defined, Gifbot will pick one at random. You can also @ your gifbot with the following
message to see all the commands it supports:

```
@gifbot help
```

## Getting

We distribute this bot as a docker image retrievable from Docker Hub under `hacklanta/gifbot`. You
could also build and run it directly on your machine.

## Running

When you run Gifbot it'll expects the following environment variables to be set when it runs:

* `SLACK_TOKEN` - The token provided by Slack for authentication.
* `DATABASE_PATH` - The path to the Bolt database for this Gifbot instance to use.

Note that if you're using the Docker distribution, you'll probably want to Bolt database to be
mounted from the host operating system or contained in a Docker volume so it doesn't get lost when
you upgrade to new versions of Gifbot.

## Building

To get the source for this project in your gopath, invoke `go get`:

```
go get -u github.com/hacklanta/gifbot
```

Then `cd` into the working directory for the project created by `go get` and invoke `go build`.
This will produce an executable for your current platform.

To build the executable for the docker image, specify that you specifically want a linux binary:

```
env GOOS=linux go build
```

You can then invoke `docker build .` to construct the actual Docker image.
