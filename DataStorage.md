# Storing data
Requirements:
- Performance
- Reproducible
- Integrity
---
# Solution

### The `raw` data is saved, that is recieved from the server, without any modifications.
Lookups are made different depending on the use-case / purpose.

Metadata look-ups can be handled by a sumerized version e.g. month.json or year.json `NOTE: day.json will be kept in the RAW format`

summerizing is handled by the aggregate function. 
### Deleted videos:

A shortcut for the video satistics, Normally the (view / like) metadata from the summerized version is deemed invalid as it may change over time, but there is one edge case, a deleted video. Deleted vidos keep thier stats indefentetly, loading the deleted.json file containg a maping from video id to deletion timestamp allows for some quick optimizations.
for example:
```go
func RequestDate(videoId int64, requestTs time.Time) (VideoData, bool) {
    if deletedTs, wasDeleted := deleted.load(video.id) ; wasDeleted && wasDeleted.After() {
	    	return metadata.Load(videoId)
    }
}
```
this is a oversimplification, though the optimisation gain is real, ram is faster than I/O overall.

## Performance
This procedure works, however it is slow due to the implementation not loading the recorded data globally for multiple requests to reuse it. This cannot be done due to the memory usage.
### The solution
We create a double linked list, containing a validity date, and the data relevant for changes over time e.g. likes, views, we can now throw away any duplicate data and assume the previous recorded state is still valid.
Then we load the data at the start of the program and use the data from RAM. this strategy does not scale for one type of video: One that is viewed daily, for years to come, however this is hard to optimize for to begin with, and if it becomes a problem, you can just split the Double linked list into time segments, such as year/month.json