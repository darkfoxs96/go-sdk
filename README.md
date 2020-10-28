# api.video Golang SDK

The [api.video](https://api.video/) web-service helps you put video on the web without the hassle. This documentation helps you use the corresponding Golang client.

## Installation
```bash
go get github.com/darkfoxs96/go-sdk
```

## Quick Start

For a more advanced usage you can checkout the rest of the documentation in the [docs directory](/docs)

```golang
package main

import (
	"fmt"
	"os"

	apivideosdk "github.com/darkfoxs96/go-sdk"
)

func main() {
    //Connect to production environment
    client := apivideosdk.NewClient(os.Getenv("API_VIDEO_KEY"))

    //Alternatively, connect to the sandbox environment for testing
    client := apivideosdk.NewSandboxClient(os.Getenv("API_VIDEO_SANDBOX_KEY"))

    //List Videos
    //First create the url options for searching
    opts := &apivideosdk.VideoOpts{
        CurrentPage: 1,
        PageSize: 25,
        SortBy:    "publishedAt",
        SortOrder: "desc",
    }

    //Then call the List endpoint with the options
    result, err := client.Videos.List(opts)
    
    if err != nil {
        fmt.Println(err)
    }

    for _, video := range result.Data {
        fmt.Printf("%s\n", video.VideoID)
        fmt.Printf("%s\n", video.Title)
    }

    //Upload a video
    //First create a container
    videoRequest := &apivideosdk.VideoRequest{
        Title: "My video title",
    }
    newVideo, err := client.Videos.Create(videoRequest)
    if err != nil {
        fmt.Println(err)
    }

    //Then upload your video to the container with the videoID
    uploadedVideo, err := client.Videos.Upload(newVideo.VideoID, "path/to/video.mp4")
    if err != nil {
        fmt.Println(err)
    }

    //And get the assets
    fmt.Printf("%s\n", uploadedVideo.Assets.Hls)
    fmt.Printf("%s\n", uploadedVideo.Assets.Iframe)
}
```