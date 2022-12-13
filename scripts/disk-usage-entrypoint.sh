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

# Run registry in background mode to be able to poll
# for the done status file and kill process after


while [ ! -f /status/done ]; do
    echo -e "-------------------------------------------------------------------------------------------------"
    
    for folder in "/root/.cache/go-build" "/home/prow/go/pkg/mod" "/home/user/.local/share/buildkit" "/var/lib/registry" "/tmp"; do
        if [ -d $folder ]; then
            du -sh $folder 2> /dev/null
        fi
    done

    for folder in "/home/prow/go/src/github.com/aws"; do
        if [ -d $folder ]; then
            echo -e "\n--------------- $folder -------------------"
            du -Sh $folder 2> /dev/null | sort -rh | head -10 
            du -sh $folder 2> /dev/null
            echo -e "--------------- $folder -------------------\n"
        fi
    done

    df -h
    echo -e "-------------------------------------------------------------------------------------------------\n"
    
    sleep 20
done
