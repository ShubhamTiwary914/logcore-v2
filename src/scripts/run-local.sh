## run and setup everything locally
## --------------------------------

#kubernetes namespaces create (if not exists)
ns="verne"
create_k8s_ns(){
    kubectl get ns $ns >/dev/null 2>&1 
    if [ $? -ne 0 ]; then
        echo "$ns namespace doesnt' exist, creating now..."
        kubectl create ns $ns
    else
        echo "$ns namespace already exists, skipping..."
    fi
}
create_k8s_ns
ns="observe"
create_k8s_ns


#pull containers


