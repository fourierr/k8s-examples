# Enhanced horizontal scaling of app controller

**Is your feature request related to a problem? Please describe.**
Due to the leader-election of the controller, when there are thousands of application CRs in the hubcluster, the new
application will be delayed for updating longer and longer time.

**Describe the solution you'd like**
We want to dispatch several deployment of app controller in hubcluster. Firstly, enabling the leader election and
set `leader-election-resource-name` with different name e.g sharding-1, sharding-2. Secondly, we add a label tag for
every application CR, e.g `app.oam.dev/sharding=sharding-1`
Thirdly, add a filter to controller's eventFilter. Only the value of label tag is equal
to `leader-election-resource-name`, the controller will handle the application CR.

The final effect is: application CR with label value of `sharding-1` will only be processed by app controller
with `leader-selection-resource-name` value of `sharding-1`



# Support to inject custom runtime context

**Is your feature request related to a problem? Please describe.**
In context definition, some runtime context information can be obtained through the context variable. 
It's real convenient for wirting cue template of component, trait or workflowstep.
However, some private business concept can't be inject into context without hacking app controller.
So we want to find a way to inject some custom context for writing cue template with convenience.

**Describe the solution you'd like**
The easy way is injecting into workload's anno e.g 
```yaml
context.oam.dev/unitID: my-unit-2
```

In `ComponentDefinition` and `TraitDefinition`, you can use context.unitID to get the value instead of via a parameter.


