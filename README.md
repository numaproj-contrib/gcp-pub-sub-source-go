
# GCP Pub/Sub Source for Numaflow

The GCP Pub/Sub Source is a custom source for [Numaflow](https://numaflow.numaproj.io/) that integrates Google Cloud Pub/Sub into Numaflow pipelines.

## Quick Start
This guide will walk you through setting up a GCP Pub/Sub source in a Numaflow pipeline.

### Prerequisites
- [Numaflow installed on your Kubernetes cluster](https://numaflow.numaproj.io/quick-start/).
- Access to a Google Cloud Platform project with Pub/Sub enabled.

### Step-by-Step Guide

1. **Create a GCP Pub/Sub Topic and Subscription**: Use the Google Cloud Console or `gcloud` CLI to create a new topic and subscription.

2. **Set Up Environment Variables**: Define environment variables such as `PROJECT_ID`, `TOPIC_ID`, `SUBSCRIPTION_ID`, `MAX_EXTENSION_PERIOD` and `PUBSUB_EMULATOR_HOST` in your Kubernetes deployment. These variables are crucial for configuring the Pub/Sub source within your pipeline.
   PROJECT_ID: This is the unique identifier for your Google Cloud Project. Every resource in Google Cloud belongs to a project. This ID is used by the Pub/Sub client to know which project's Pub/Sub resources it should interact with.

TOPIC_ID: The ID of the Pub/Sub topic from which messages are published. Topics in Pub/Sub are named resources to which messages are sent by publishers.

SUBSCRIPTION_ID: The ID of the Pub/Sub subscription through which messages are consumed. Subscriptions in Pub/Sub represent a named resource mapping to a topic, through which subscribers receive messages published to the topic.

MAX_EXTENSION_PERIOD: This is typically used to set the maximum period for which the Pub/Sub client will automatically extend the acknowledgment deadline for each message. This setting is crucial for handling message processing that might take longer than the default acknowledgment deadline.

PUBSUB_EMULATOR_HOST: This is used when developing and testing applications locally, using the Pub/Sub emulator rather than connecting to the actual Pub/Sub service. This variable should hold the address (host and port) of the running Pub/Sub emulator.For Production Keep it blank.

3. **Deploy a Numaflow Pipeline with GCP Pub/Sub Source**: Use the following Kubernetes manifest template, modifying the environment variables as necessary:

   ```yaml
   apiVersion: numaflow.numaproj.io/v1alpha1
   kind: Pipeline
   metadata:
     name: pubsub-source-pipeline
   spec:
     vertices:
       - name: pubsub-source
         source:
           udsource:
             container:
               image: "quay.io/numaio/numaflow-go/gcp-pubsub-source-go:latest"
               env:
                 - name: PROJECT_ID
                   value: "your-gcp-project-id"
                 - name: TOPIC_ID
                   value: "your-pubsub-topic-id"
                 - name: SUBSCRIPTION_ID
                   value: "your-pubsub-subscription-id"
                 - name: PUBSUB_EMULATOR_HOST
                   value: "gcloud-pubsub.numaflow-system.svc.cluster.local:8681"
       - name: log-sink
         sink:
           log: {}
     edges:
       - from: pubsub-source
         to: log-sink
   ```

   Apply it with `kubectl apply -f pubsub-source-pipeline.yaml`.

4. **Verify the Pipeline**: Check its status with `kubectl get pipeline pubsub-source-pipeline`.

5. **Publish Messages to the GCP Pub/Sub Topic**: Use the `gcloud` CLI to publish messages:

   ```bash
   gcloud pubsub topics publish <your-pubsub-topic-id> --message "Your message here"
   ```

6. **Verify the Log Sink**: Check the logs of the `log-sink` vertex with `kubectl logs <log-sink-pod-name>` to confirm message reception.

7. **Clean Up**:
    - Delete the pipeline: `kubectl delete pipeline pubsub-source-pipeline`.
    - Delete the Pub/Sub topic and subscription:

      ```bash
      gcloud pubsub topics delete <topic-id>
      gcloud pubsub subscriptions delete <subscription-id>
      ```

## Debugging

Enable debugging by setting `DEBUG=true` in the GCP Pub/Sub source container's environment variables. For further details, see the [Numaflow Debugging Guide](https://numaflow.numaproj.io/development/debugging/).

## Using with Managed K8s
If you're using any managed k8 cluster other than GCP, like Amazon EKS , you'll need to manage credentials differently since Google Cloud's automatic workload identity features aren't available. Follow these steps to set up your application:

Create a Google Cloud service account with the necessary roles for Pub/Sub and Cloud Monitoring.

Download the service account JSON key file.

Create a Kubernetes Secret in your EKS cluster with the JSON key file:


``kubectl create secret generic google-cloud-key --from-file=key.json=/path/to/service-account-file.json``

Configure your pod to use this secret by mounting it as a volume. Here's an example pod configuration:

```yaml
apiVersion: v1
kind: Pod
metadata:
name: my-pod
spec:
containers:
- name: my-container
  image: my-image
  env:
   - name: GOOGLE_APPLICATION_CREDENTIALS
     value: /var/secrets/google/key.json
     volumeMounts:
   - name: google-cloud-key
     mountPath: /var/secrets/google
     volumes:
- name: google-cloud-key
  secret:
  secretName: google-cloud-key
```

#### The GOOGLE_APPLICATION_CREDENTIALS environment variable points to the file where the secret is mounted in the container.


Ensure the service account has the necessary permissions to perform operations on Google Cloud resources.


## Additional Resources

For more information, visit the [Numaflow Documentation](https://numaflow.numaproj.io/) and [Google Cloud Pub/Sub Documentation](https://cloud.google.com/pubsub/docs).