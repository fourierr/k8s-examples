
## Code Map
### /apis 
Package apis contains all api types of KubeVela.

#### /apis/core.oam.dev
Package core.oam.dev contains API Schema definitions for the core.oam.dev API group.

#### /apis/standard.oam.dev
Package standard.oam.dev contains API Schema definitions for the standard.oam.dev API group.

#### /apis/types
Package types contains auxiliary definitions for the API group.

### /charts
Used as a Helm Chart.

### /cmd
Place for initialization code.

#### /cmd/apiserver
Place for `vela-apiserver` initialization code.

#### /cmd/core
Place for `vela-core` initialization code.

#### /cmd/plugin
Place for command line interface `vela` initialization code.

### /contribute
Place for Contributor / Developer guide.

### /design
The design for api, platform, vela-cli and vela-core.

### /docs
Place for api docs and examples of feature.

### /e2e and /test
Code that does the e2e-tests.

### /makefiles
Used as a collection of shortcuts, e.g `make build` or `make reviewable`

### /pkg
Main place for the go code.

#### /pkg/addon
Addon related code.

#### /pkg/apiserver
`vela-apiserver` related code.

#### /pkg/appfile
Code for appfile that is an important go structure of `vela-core`.

#### /pkg/auth
The authorization for KubeVela.

#### /pkg/client
The client for controller_client and delegating_client.

#### /pkg/controller
Custom resource controllers for core.oam.dev and standard.oam.dev API group.

#### /pkg/cue
Cue file related code, contains validation and generation etc.

#### /pkg/monitor
It contains a context that supports fork and commit like trace span.

#### /pkg/multicluster
Code for multiple cluster delivery.

#### /pkg/policy
Policy related code, contains topology, override and envbinding.

#### /pkg/resourcekeeper
ResourceKeeper handler for dispatching、 deleting resources and keeping resources up-to-date.

#### /pkg/stdlib
The lib for cue operators.

#### /pkg/utils
Common utils for KubeVela.

#### /pkg/velaql
Place for VelaQL. It is a resource query language for KubeVela, used to query status of any extended resources in application-level.

#### /pkg/workflow
Code for workflow.

### /references
Place for `vela-cli` reference implementation.

### /vela-templates
This is the place for hold built-in CUE templates for Vela Core and Registry.