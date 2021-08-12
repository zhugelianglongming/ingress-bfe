# 配置优先级
## 路由配置冲突
当用户的ingress配置最终生成相同的路由规则的情况下（Host、Path、Header/Cookie完全相同），BFE-Ingress将按照_**创建时间优先**_的原则使用先配置的路由规则。
对于因路由冲突导致的配置生成失败，ingress-controller会落下响应的错误日志。
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress-A
spec:
  rules:
  - host: example.foo.com
    http:
      paths:
      - path: /foo
        pathType: Prefix
        backend:
          service:
            name: service1
            port:
              number: 80
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress-B
spec:
  rules:
  - host: example.foo.com
    http:
      paths:
      - path: /foo
        pathType: Prefix
        backend:
          service:
            name: service2
            port:
              number: 80
#其中ingress-A先于配置ingress-B创建，则最终仅生效ingress-A。
```
## 跨namespace冲突
下面将说明BFE-Ingress在监听多个namespace且这多个namespace之间存在配置冲突的情况的处理方式。

- namespace之间存在路由规则冲突？
   - 按照路由冲突章节的描述来处理。
- namespace之间存在命名冲突，包括但不限于ingress命名，service命名？
   - ingress命名在controller内部会按照${namespace}/${ingress}的方式来配置，因此命名没有影响；
   - service的场景则按照：仅匹配ingress资源所在namespace下的资源；例如namespaceA下的ingress仅能将namespaceA下的service当做后端生成配置；
## 状态回写
当前Ingress的合法性是在配置生效的过程才能感知，是一个异步过程。为了能给用户反馈当前Ingress是否生效，BFE-Ingress会将Ingress的实际生效状态回写到Ingress的一个Annotation当中。
**BFE-Ingress状态Annotation定义如下：**
```yaml
#bfe.ingress.kubernetes.io/bfe-ingress-status为BFE-Ingress预留的Annotation key，
#用于BFE-Ingress回写状态
# status; 表示当前ingress是否合法， 取值为：success -> ingress合法， error -> ingress不合法
# message; 当ingress不合法的情况下，message记录错误详细原因。
bfe.ingress.kubernetes.io/bfe-ingress-status: {"status": "", "message": ""}
```
**下面是BFE-Ingress状态回写的示例：**
`Ingress1`和`Ingress2`的路由规则完全一样(`Host:example.net, Path:/bar`)。
```yaml
kind: Ingress
apiVersion: extensions/v1beta1
metadata:
  name: "ingress1"
  namespace: production
spec:
  rules:
    - host: example.net
      http:
        paths:
          - path: /bar
            backend:
              serviceName: service1
              servicePort: 80
---
kind: Ingress
apiVersion: extensions/v1beta1
metadata:
  name: "ingress2"
  namespace: production
spec:
  rules:
    - host: example.net
      http:
        paths:
          - path: /foo
            backend:
              serviceName: service2
              servicePort: 80
```
根据路由冲突配置规则，`Ingress1`将生效，而`Ingress2`将被忽略。状态回写后，`Ingress1`的状态为success，而`Ingress2`的状态为fail。
```yaml
kind: Ingress
apiVersion: extensions/v1beta1
metadata:
  name: "ingress1"
  namespace: production
  annotations:
    bfe.ingress.kubernetes.io/bfe-ingress-status: {"status": "success", "message": ""}
spec:
  rules:
    - host: example.net
      http:
        paths:
          - path: /bar
            backend:
              serviceName: service1
              servicePort: 80
---
kind: Ingress
apiVersion: extensions/v1beta1
metadata:
  name: "ingress2"
  namespace: production
  annotations:
    bfe.ingress.kubernetes.io/bfe-ingress-status: |
    	{"status": "fail", "message": "conflict with production/ingress1"}
spec:
  rules:
    - host: example.net
      http:
        paths:
          - path: /foo
            backend:
              serviceName: service2
              servicePort: 80
```