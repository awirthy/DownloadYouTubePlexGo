# /etc/cron.d/ytdl
# 
# go run TEST-Go.go
go run /opt/DownloadYouTubePlexGo/DownloadYouTubePlexGo-1.00/DownloadYouTubePlexGo.go  >> /proc/1/fd/1;
echo "DONE"  >> /proc/1/fd/1;