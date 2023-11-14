
# GCP Pub/Sub Source for Numaflow

The GCP Pub/Sub Source is a custom source for [Numaflow](https://numaflow.numaproj.io/) that integrates Google Cloud Pub/Sub into Numaflow pipelines.

## Quick Start
This guide will walk you through setting up a GCP Pub/Sub source in a Numaflow pipeline.

### Prerequisites
- [Numaflow installed on your Kubernetes cluster](https://numaflow.numaproj.io/quick-start/).
- Access to a Google Cloud Platform project with Pub/Sub enabled.

### Step-by-Step Guide

1. **Create a GCP Pub/Sub Topic and Subscription**: Use the Google Cloud Console or `gcloud` CLI to create a new topic and subscription.

2. **Set Up Environment Variables**: Define environment variables such as `PROJECT_ID`, `TOPIC_ID`, `SUBSCRIPTION_ID`, and `PUBSUB_EMULATOR_HOST` in your Kubernetes deployment. These variables are crucial for configuring the Pub/Sub source within your pipeline.

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

## Additional Resources

For more information, visit the [Numaflow Documentation](https://numaflow.numaproj.io/) and [Google Cloud Pub/Sub Documentation](https://cloud.google.com/pubsub/docs).