INIT_OPTS=init \
	--workload-config .operator-builder/workload.yaml \
   	--repo github.com/nukleros/supporting-services-operator \
	--controller-image nukleros/supporting-services-operator:latest \
    --skip-go-version-check

CREATE_OPTS=create api \
	--workload-config .operator-builder/workload.yaml \
	--controller \
	--resource

operator-clean:
	rm -rf ../*

operator-init:
	cd .. && operator-builder $(INIT_OPTS)

operator-create:
	cd .. && operator-builder $(CREATE_OPTS)
