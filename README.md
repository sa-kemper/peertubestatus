# golangGetTextTest
This is a simple golang webserver that localizes it's pages to the requested language (if available) in a concurrently safe manner, this project is used mainly for testing purposes but may be used as a starting point of a real project.
# peertubestats
peertubestats is a program written in golang. It obtains statistics from a running peertube instance, and saves them in a raw format, so that any bugs that this program has are not affecting the data collected. The collected data is added to the custom save strategy. The custom strategy invovles the metadata of every video, mapped from video id to data. The other data saved is frequently updated data such as views and likes. This frequently changing data is saved in a double linked list format, where the key is the date of the data. by moving down the linked list you can obtain more current data. There is no duplicate data in this double linked list, it only tracks changes of the data. Each video gets its own file, and the stats are tracked separately, this allows this program too scale to millions of videos. 

The scope of this project was to obtain stats and provide a static site for the obtained stats, however due to usability concerns we expanded the logic with a small backend.
# Installation
To install the service, just compile the binaries and run them to your liking. However, there are a few convenience options available.
## Compilation:
We assume you use a posix compliant operating system, you have cloned the source code to a directory of your choosing and have golang 1.25.2 installed.

There are no required build flags, however there is no need for a lot of debug information in production, the following commands will build all executables in our suggested way:
```shell
go build -ldflags="-s -w" ./cmd/CronSaveStats # A utility used as a cron service to save the current peertube data.
go build -ldflags="-s -w" ./cmd/peertubeExportStat # A utility for generating a report of every video into a static html files.
go build -ldflags="-s -w" ./cmd/peertubestats # A statistics go http server with search and interactivity. Should be used in combination with CronSaveStats 
```

Neither the peertubeExportStat nor the peertubestats http service obtain any data from the peertube instance. use the CronSaveStats utility for that. `NOTE: A restart of the peertubestats application should be done so the data is reloaded. A simple solution for this is a cronjob that restarts the unit.`

## Sample installation, Step by step.
```shell
cd "/opt"
git clone https://github.com/sa-kemper/peertubestats
cd peertubestats

go build -ldflags="-s -w" ./cmd/CronSaveStats
go build -ldflags="-s -w" ./cmd/peertubeExportStat
go build -ldflags="-s -w" ./cmd/peertubestats

systemctl edit --force --full peertubestats.service
```
 - Now enter the following with your editor:
```systemd
[Unit]
Description=Peertube stats web service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/peertubestats
ExecStart=/opt/peertubestats/peertubestats

[Install]
WantedBy=multi-user.target
```
 - Write and quit the editor
```shell
cp dotenv .env # You should really edit this. take a look at each Usage information for the respective executable, they are all compatible, but some use more and or different config flags then others.
$EDITOR .env
$EDITOR /etc/crontab
``` 
 - Now insert the following two Cron entries 
```cron
1 * * * * root /opt/peertubestats/CronSaveStats
10 * * * * root /usr/bin/systemctl restart peertubsestats
```
- You now have saving and displaying of the peertube stats data.

For more installation documentation review the [After Basic Install](AfterBasics.Install.md) guide.


---
### Override content.
| Override file type      | override location                                                                                                                                                                                 |
|-------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Override html templates | .gohtml files into the work directory/TemplateOverride/                                                                                                                                           |
| Override css files      | place a folder named "static" in the work dir of the executeable, inside that folder make another folder named css, place your css files here. the main one that is allways included is `style.css` |

# Dependencies
From the backend side of things we are only depend on gotext (gettext implemented for golang), this causes a indirect dependency on the standard text library of golang.

The frontend is dependent on [Charts.css](https://chartscss.org/) and [Font Awesome](https://fontawesome.com/). These could be hosted by peertubestats, by overriding the template files and pointing them to the static folder, however making this process easier is not part of the scope at the moment.