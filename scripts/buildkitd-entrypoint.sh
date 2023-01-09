# Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Run buildkitd in background mode to be able to poll
# for the done status file and kill process after

# if running on the postsubmit cluster allow more cache space
KEEP_STORAGE=20000
if [ -n "${FARGATE_PROFILE_NAME}" ]; then
	KEEP_STORAGE=5000
fi

rootlesskit \
	buildkitd \
	--addr=unix:///run/buildkit/buildkitd.sock \
	--oci-worker-no-process-sandbox \
	--oci-worker-platform=linux/amd64 \
	--oci-worker-platform=linux/arm64 \
	--oci-worker-gc \
	--oci-worker-gc-keepstorage $KEEP_STORAGE \
	&
pid=$!
while [ ! -f /status/done ]
do
  sleep 5
done
kill $pid

