// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package list

import (
	"context"
	"testing"

	"github.com/Azure/go-autorest/autorest/to"
	"github.com/golang/mock/gomock"
	v1 "github.com/project-radius/radius/pkg/armrpc/api/v1"
	"github.com/project-radius/radius/pkg/cli/clients"
	"github.com/project-radius/radius/pkg/cli/connections"
	"github.com/project-radius/radius/pkg/cli/framework"
	"github.com/project-radius/radius/pkg/cli/objectformats"
	"github.com/project-radius/radius/pkg/cli/output"
	"github.com/project-radius/radius/pkg/cli/workspaces"
	"github.com/project-radius/radius/pkg/corerp/api/v20220315privatepreview"
	"github.com/project-radius/radius/test/radcli"
	"github.com/stretchr/testify/require"
)

func Test_CommandValidation(t *testing.T) {
	radcli.SharedCommandValidation(t, NewCommand)
}

func Test_Validate(t *testing.T) {
	configWithWorkspace := radcli.LoadConfigWithWorkspace(t)
	testcases := []radcli.ValidateInput{
		{
			Name:          "Valid List Command",
			Input:         []string{},
			ExpectedValid: true,
			ConfigHolder: framework.ConfigHolder{
				ConfigFilePath: "",
				Config:         configWithWorkspace,
			},
		},
		{
			Name:          "List Command with fallback workspace",
			Input:         []string{"-e", "my-env", "-g", "my-env"},
			ExpectedValid: false,
			ConfigHolder: framework.ConfigHolder{
				ConfigFilePath: "",
				Config:         radcli.LoadEmptyConfig(t),
			},
		},
		{
			Name:          "List Command with too many args",
			Input:         []string{"foo", "bar"},
			ExpectedValid: false,
			ConfigHolder: framework.ConfigHolder{
				ConfigFilePath: "",
				Config:         configWithWorkspace,
			},
		},
	}
	radcli.SharedValidateValidation(t, NewCommand, testcases)
}

func Test_Run(t *testing.T) {
	t.Run("List recipes linked to the environment", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			ctrl := gomock.NewController(t)

			envResource := v20220315privatepreview.EnvironmentResource{
				ID:       to.StringPtr("/planes/radius/local/resourcegroups/kind-kind/providers/applications.core/environments/kind-kind"),
				Name:     to.StringPtr("kind-kind"),
				Type:     to.StringPtr("applications.core/environments"),
				Location: to.StringPtr(v1.LocationGlobal),
				Properties: &v20220315privatepreview.EnvironmentProperties{
					Recipes: map[string]*v20220315privatepreview.EnvironmentRecipeProperties{
						"cosmosDB": {
							LinkType:     to.StringPtr("Applications.Link/mongoDatabases"),
							TemplatePath: to.StringPtr("testpublicrecipe.azurecr.io/bicep/modules/mongodatabases:v1"),
						},
					},
				},
			}
			recipes := []EnvironmentRecipe{
				{
					Name:         "cosmosDB",
					LinkType:     "Applications.Link/mongoDatabases",
					TemplatePath: "testpublicrecipe.azurecr.io/bicep/modules/mongodatabases:v1",
				},
			}

			appManagementClient := clients.NewMockApplicationsManagementClient(ctrl)
			appManagementClient.EXPECT().
				GetEnvDetails(gomock.Any(), gomock.Any()).
				Return(envResource, nil).Times(1)

			outputSink := &output.MockOutput{}

			runner := &Runner{
				ConnectionFactory: &connections.MockFactory{ApplicationsManagementClient: appManagementClient},
				Output:            outputSink,
				Workspace:         &workspaces.Workspace{},
				Format:            "table",
			}

			err := runner.Run(context.Background())
			require.NoError(t, err)

			expected := []interface{}{
				output.FormattedOutput{
					Format:  "table",
					Obj:     recipes,
					Options: objectformats.GetEnvironmentRecipesTableFormat(),
				},
			}
			require.Equal(t, expected, outputSink.Writes)
		})
	})
}
