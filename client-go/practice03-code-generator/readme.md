1、获取code-generator的代码，并切换到[v0.23.3 ](https://github.com/kubernetes/code-generator/releases/tag/v0.23.3)的tag上
git checkout v0.23.3

2、编译项目，安装代码生成工具，这里我们只安装我们接下来会用到的工具,
go install code-generator/cmd/{client-gen,lister-gen,informer-gen,deepcopy-gen}

3、使用工具code-generator/generate-groups.sh
code-generator/generate-groups.sh all MOD_NAME/pkg/generated MOD_NAME/pkg/apis foo.example.com:v1 --output-base MOD_DIR/..  --go-header-file "code-generator/hack/boilerplate.go.txt"

4、通常直接包级别标记,定义在doc.go

```
// +k8s:deepcopy-gen=package
// +groupName=foo.example.com
package v1
```

备注：2、3选一种即可