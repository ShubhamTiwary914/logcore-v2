#context: seperated phases for testing components seperately
#GKE phase: GKE and VPC network for VerneMQTT pods and listener
#pipeline phase: streampipeline with pubsub, beamSDK & bigtable (managed services)


module "gke-phase" {
    source = "./gke/"
    region = var.region
    logcore_serviceacc = var.logcore_serviceacc
    project_id = var.project_id
}

module "pipeline-phase" {
    source = "./pipeline/"  
    region = var.region
    logcore_serviceacc = var.logcore_serviceacc
    project_id = var.project_id
}