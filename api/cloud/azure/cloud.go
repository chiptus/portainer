// Copyright © 2019 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package azure

import (
	"time"

	"github.com/Azure/azure-sdk-for-go/services/apimanagement/mgmt/2021-08-01/apimanagement"
	"github.com/Azure/azure-sdk-for-go/services/authorization/mgmt/2020-10-01/authorization"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2021-12-01/compute"
	cs "github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-02-01/containerservice"
	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2022-02-01/containerservice"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2018-10-01/network"
	"github.com/Azure/azure-sdk-for-go/services/preview/monitor/mgmt/2018-09-01/insights"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-02-01/resources"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2021-01-01/subscriptions"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
)

const userAgent = "Banzai Pipeline"

// CloudConnection represents an authenticated Azure cloud connection
type CloudConnection struct {
	env    azure.Environment
	creds  Credentials
	client autorest.Client
	cache  struct {
		authorizationBaseClient    *authorization.BaseClient
		computeBaseClient          *compute.BaseClient
		managedClusterBaseClient   *containerservice.BaseClient
		containerServiceBaseClient *cs.BaseClient
		insightsBaseClient         *insights.BaseClient
		networkBaseClient          *network.BaseClient
		resourcesBaseClient        *resources.BaseClient
		subscriptionsBaseClient    *subscriptions.BaseClient
		apiManagementBaseClient    *apimanagement.BaseClient
		resourceSKUsClient         *compute.ResourceSkusClient
	}
}

// NewCloudConnection returns a new CloudConnection instance
func NewCloudConnection(env *azure.Environment, creds *Credentials) (*CloudConnection, error) {
	cc := &CloudConnection{
		env:    *env,
		creds:  *creds,
		client: autorest.NewClientWithUserAgent(userAgent),
	}

	cc.client.PollingDuration = 30 * time.Minute

	var err error
	cc.client.Authorizer, err = GetAuthorizer(&creds.ServicePrincipal, env)
	if err != nil {
		return nil, err
	}
	return cc, nil
}

// GetSubscriptionID returns the subscription ID used to connect to the cloud
func (cc *CloudConnection) GetSubscriptionID() string {
	return cc.creds.SubscriptionID
}
