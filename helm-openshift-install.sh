#!/bin/bash

#
# Copyright Red Hat
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

if [ $# -eq 0 ]; then
    echo "Usage: $0 <arguments>"
    exit 1
fi

START_TIME=$(date +%s)
TIMEOUT_SEC=25 
SPINUP_SEC=60 # Helm Upgrade can have unexpected results if it is run while the cluster is initially starting

# Run the install without the generated Openshift Route being passed to Registry Viewer fqdn as it does not exist yet
helm install devfile-registry deploy/chart/devfile-registry --set global.isOpenShift=true "$@"

while true; do
    ROUTE=$(oc get route devfile-registry -o jsonpath='{.spec.host}' 2>/dev/null)
    if [ -n "$ROUTE" ]; then
        echo Domain found: "$ROUTE"
        break
    fi
    new_time=$(date +%s)
    elapsed_time=$((new_time - START_TIME))
    if [ $elapsed_time -ge $TIMEOUT_SEC ]; then
        echo "TIMEOUT: Domain not found."
        exit 1
    fi
    sleep 1
done

# Allow deployment to start occuring before the upgrade
sleep $SPINUP_SEC

# Run upgrade with the new variable to set fqdn of the Registry Viewer
helm upgrade devfile-registry deploy/chart/devfile-registry --reuse-values --set global.route.domain=$ROUTE 
