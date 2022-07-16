import radius as radius

param magpieimage string = 'radiusdev.azurecr.io/magpiego:latest'
param environment string

resource app 'Applications.Core/applications@2022-03-15-privatepreview' = {
  name: 'corerp-mechanics-redeploy-withtwoseparateresource'
  location: 'global'
  properties: {
    environment: environment
  }
}

resource b 'Applications.Core/containers@2022-03-15-privatepreview' = {
  name: 'corerp-mechanics-redeploy-withanotherresource-b'
  location: 'global'
  properties: {
    application: app.id
    container: {
      image: magpieimage
    }
  }
}