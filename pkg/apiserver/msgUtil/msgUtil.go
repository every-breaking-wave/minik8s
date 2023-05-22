package msgutil

import (
	"encoding/json"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
	"miniK8s/pkg/entity"
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/message"
	"miniK8s/util/stringutil"
)

type MsgUtil struct {
	Publisher *message.Publisher
}

var ServerMsgUtil *MsgUtil

func init() {
	// 初始化消息队列
	newPublisher, err := message.NewPublisher(message.DefaultMsgConfig())
	if err != nil {
		panic(err)
	}
	ServerMsgUtil = &MsgUtil{
		Publisher: newPublisher,
	}
}

// 发布消息的组件
func PublishMsg(queueName string, msg []byte) error {
	return ServerMsgUtil.Publisher.Publish(queueName, message.ContentTypeJson, msg)
}

// 发布消息的组件函数
func PublishRequestNodeScheduleMsg(pod *apiObject.PodStore) error {
	resourceURI := stringutil.Replace(config.PodSpecURL, config.URL_PARAM_NAME_PART, pod.GetPodName())
	resourceURI = stringutil.Replace(resourceURI, config.URL_PARAM_NAMESPACE_PART, pod.GetPodNamespace())
	podJson, err := json.Marshal(pod)
	if err != nil {
		k8log.ErrorLog("msgutil", "json marshal pod failed")
		return err
	}
	message := message.Message{
		Type:         message.RequestSchedule,
		Content:      string(podJson),
		ResourceURI:  resourceURI,
		ResourceName: pod.GetPodName(),
	}

	jsonMsg, err := json.Marshal(message)

	if err != nil {
		return err
	}

	return PublishMsg(NodeSchedule, jsonMsg)
}

func PublishUpdateService(serviceUpdate *entity.ServiceUpdate) error {
	resourceURI := stringutil.Replace(config.ServiceSpecURL, config.URL_PARAM_NAME_PART, serviceUpdate.ServiceTarget.GetName())
	resourceURI = stringutil.Replace(resourceURI, config.URL_PARAM_NAMESPACE_PART, serviceUpdate.ServiceTarget.GetNamespace())

	jsonBytes, err := json.Marshal(serviceUpdate)
	if err != nil {
		return err
	}

	message := message.Message{
		Type:         message.CREATE,
		Content:      string(jsonBytes),
		ResourceURI:  resourceURI,
		ResourceName: serviceUpdate.ServiceTarget.GetName(),
	}

	jsonMsg, err := json.Marshal(message)

	if err != nil {
		return err
	}

	return PublishMsg(ServiceUpdate, jsonMsg)
}

func PublishUpdateEndpoints(endpointUpdate *entity.EndpointUpdate) error {
	resourceURI := stringutil.Replace(config.ServiceSpecURL, config.URL_PARAM_NAME_PART, endpointUpdate.ServiceTarget.Service.GetName())
	resourceURI = stringutil.Replace(resourceURI, config.URL_PARAM_NAMESPACE_PART, endpointUpdate.ServiceTarget.Service.GetNamespace())

	jsonBytes, err := json.Marshal(endpointUpdate)
	if err != nil {
		return err
	}

	message := message.Message{
		Type:         message.CREATE,
		Content:      string(jsonBytes),
		ResourceURI:  resourceURI,
		ResourceName: endpointUpdate.ServiceTarget.Service.GetName(),
	}

	jsonMsg, err := json.Marshal(message)

	if err != nil {
		return err
	}

	return PublishMsg(EndpointUpdate, jsonMsg)
}

func PublishUpdatePod(podUpdate *entity.PodUpdate) error {
	resourceURI := stringutil.Replace(config.PodSpecURL, config.URL_PARAM_NAME_PART, podUpdate.PodTarget.GetPodName())
	resourceURI = stringutil.Replace(resourceURI, config.URL_PARAM_NAMESPACE_PART, podUpdate.PodTarget.GetPodNamespace())

	jsonBytes, err := json.Marshal(podUpdate)
	if err != nil {
		return err
	}

	message := message.Message{
		Type:         podUpdate.Action,
		Content:      string(jsonBytes),
		ResourceURI:  resourceURI,
		ResourceName: podUpdate.PodTarget.GetPodName(),
	}

	jsonMsg, err := json.Marshal(message)

	if err != nil {
		return err
	}

	// 发送给pod所在的Node监听的podUpdate消息队列
	return PublishMsg(PodUpdateWithNode(podUpdate.PodTarget.Spec.NodeName), jsonMsg)
}

func PublishDeletePod(pod *apiObject.PodStore) error {
	resourceURI := stringutil.Replace(config.PodSpecURL, config.URL_PARAM_NAME_PART, pod.GetPodName())
	resourceURI = stringutil.Replace(resourceURI, config.URL_PARAM_NAMESPACE_PART, pod.GetPodNamespace())
	jsonBytes, err := json.Marshal(pod)
	if err != nil {
		k8log.ErrorLog("msgutil", "json marshal pod failed")
		return err
	}

	message := message.Message{
		Type:         message.DELETE,
		Content:      string(jsonBytes),
		ResourceURI:  resourceURI,
		ResourceName: pod.GetPodName(),
	}

	jsonMsg, err := json.Marshal(message)

	if err != nil {
		return err
	}

	// 发送给pod所在的Node监听的podUpdate消息队列
	return PublishMsg(PodUpdateWithNode(pod.Spec.NodeName), jsonMsg)
}

// 接受的消息是job的metadata
// 这里资源的URL是job的spec的URL，而不是文件的URL
// 但是最终处理的时候会检查两者都存在的时候才会创建Pod，去执行任务
func PublishUpdateJobFile(jobMeta *apiObject.Basic) error {
	resourceURI := stringutil.Replace(config.JobSpecURL, config.URL_PARAM_NAMESPACE_PART, jobMeta.Metadata.Namespace)
	resourceURI = stringutil.Replace(resourceURI, config.URL_PARAM_NAME_PART, jobMeta.Metadata.Name)

	jsonBytes, err := json.Marshal(jobMeta)

	if err != nil {
		k8log.ErrorLog("msgutil", "json marshal job failed")
		return err
	}

	message := message.Message{
		Type:         message.UPDATE,
		Content:      string(jsonBytes),
		ResourceURI:  resourceURI,
		ResourceName: jobMeta.Metadata.Name,
	}

	jsonMsg, err := json.Marshal(message)

	if err != nil {
		return err
	}

	return PublishMsg(JobUpdate, jsonMsg)
}

func PublishUpdateDns(dnsUpdate *entity.DnsUpdate) error {
	resourceURI := stringutil.Replace(config.DnsSpecURL, config.URL_PARAM_NAMESPACE_PART, dnsUpdate.DnsTarget.Metadata.Namespace)
	resourceURI = stringutil.Replace(resourceURI, config.URL_PARAM_NAME_PART, dnsUpdate.DnsTarget.Metadata.Name)

	jsonBytes, err := json.Marshal(dnsUpdate)

	if err != nil {
		k8log.ErrorLog("msgutil", "json marshal dns failed")
		return err
	}

	message := message.Message{
		Type:         dnsUpdate.Action,
		Content:      string(jsonBytes),
		ResourceURI:  resourceURI,
		ResourceName: dnsUpdate.DnsTarget.Metadata.Name,
	}

	jsonMsg, err := json.Marshal(message)

	if err != nil {
		k8log.ErrorLog("msgutil", "json marshal dns failed")
		return err
	}

	return PublishMsg(DnsUpdate, jsonMsg)
}

func PubelishUpdateHost(hostUpdate []string) error {
	jsonBytes, err := json.Marshal(hostUpdate)

	if err != nil {
		k8log.ErrorLog("msgutil", "json marshal host failed")
		return err
	}

	// 创建一个空字符串

	message := message.Message{
		Type:         message.UPDATE,
		Content:      string(jsonBytes),
		ResourceURI:  "",
		ResourceName: "Host",
	}

	jsonMsg, err := json.Marshal(message)

	if err != nil {
		k8log.ErrorLog("msgutil", "json marshal host failed")
		return err
	}

	return PublishMsg(HostUpdate, jsonMsg)
}