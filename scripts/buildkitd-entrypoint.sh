# Run buildkitd in background mode to be able to poll
# for the done status file and kill process after
rootlesskit \
	buildkitd \
	--addr=unix:///run/buildkit/buildkitd.sock \
	--oci-worker-no-process-sandbox \
	--oci-worker-platform=linux/amd64 \
	--oci-worker-platform=linux/arm64 \
	&
pid=$!
while [ ! -f /status/done ]
do
  sleep 5
done
kill $pid
