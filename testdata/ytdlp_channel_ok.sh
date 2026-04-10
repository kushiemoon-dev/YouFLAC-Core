#!/bin/sh
# Fake yt-dlp for testing. Emits NDJSON to stdout.
for i in 1 2 3 4 5 6 7 8; do
  printf '{"id":"longform%d","title":"Long Video %d","duration":400,"upload_date":"20240101","thumbnail":"https://i.ytimg.com/vi/longform%d/hqdefault.jpg"}\n' $i $i $i
done
# 2 shorts (duration < 60)
printf '{"id":"short1","title":"Short 1","duration":30,"upload_date":"20240101","thumbnail":""}\n'
printf '{"id":"short2","title":"Short 2","duration":45,"upload_date":"20240101","thumbnail":""}\n'
