# Summary

This document includes the road map features planned for operator-builder.  It includes
**major** features and is not intended to be used to include all future plans.

Please see all [issues](https://github.com/nukleros/operator-builder/issues) for details about 
any bugs and features:

1. Complex Data Types - currently operator-builder only supports flat data types used in 
markers, such as strings, integers and booleans.  However, there are times where arrays 
are going to be needed in order to inject into the destination resources, mainly the 
[]string data type.  Currently, the target data type is [[]string](https://github.com/nukleros/operator-builder/issues/81)
but other data types such as [[]map](https://github.com/nukleros/operator-builder/issues/11) and 
[maps](https://github.com/nukleros/operator-builder/issues/10) to be considered for the future.

2. Webhook - add ability to inject [webhook validation](https://github.com/nukleros/operator-builder/issues/3) into the
generated code.  This may also be able to be accomplished with [CRD CEL](https://github.com/kubernetes/enhancements/blob/master/keps/sig-api-machinery/2876-crd-validation-expression-language/README.md) 
language.
