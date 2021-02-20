# Dusty

I have a small AKS cluster and I'm cheap.

To make sure that I don't flood my container registry with images and end up paying for extra storage, I built this quick and dirty utility.

It's written in Go because I wanted to play with the language - it's not elegant and there is likely room for improvement. Issues and PRs welcome.

## Installation

Clone the repo and `go build`

## Usage

> (Assuming Azure) Configure a Service Principal with Contributor access to your Container Registry.

There are two ways to use Dusty:

### CLI

```sh
❯ ./dusty -h
Usage of ./dusty:
  -d	Perform a dry run
  -l int
    	Number of tags you'd like to keep for each image (shorthand)
  -limit int
    	Number of tags you'd like to keep for each image
  -p string
    	Client ID/User for authentication (shorthand)
  -password string
    	Secret / Password for authentication
  -r string
    	Name of your Azure Container Registry (shorthand)
  -registry string
    	Name of your Azure Container Registry
  -u string
    	Client ID / User for authentication (shorthand)
  -user string
    	Client ID / User for authentication
  -v	Enable verbose logging
```

I'd recommend performing a dry run with the `-d` flag before an actual run to ensure you're happy. I won't be held responsible for cleaning out your registry.

### Docker

You can deploy Dusty anywhere you can place a Docker container (kubernetes, perhaps).

Instead of passing command-line arguments, you can instead set some environment variables:

|Environment Variable Name | Description |
|:---|:---|
|DUSTY_REG_NAME|Container registry name or URL|
|DUSTY_CLIENT_ID|Client ID|
|DUSTY_CLIENT_SECRET|Client Secret/Password|
|DUSTY_TAG_LIMIT|How many images you'd like to remain in each repository|
