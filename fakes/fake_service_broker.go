package fakes

import (
	"context"
	"errors"
	"reflect"
	"slices"

	"code.cloudfoundry.org/brokerapi/v13"
	"code.cloudfoundry.org/brokerapi/v13/domain"
	"code.cloudfoundry.org/brokerapi/v13/domain/apiresponses"
)

type FakeServiceBroker struct {
	ProvisionedInstances map[string]brokerapi.ProvisionDetails

	InstanceFetchDetails domain.FetchInstanceDetails
	UpdateDetails        brokerapi.UpdateDetails
	DeprovisionDetails   brokerapi.DeprovisionDetails

	DeprovisionedInstanceIDs []string
	UpdatedInstanceIDs       []string
	GetInstanceIDs           []string

	BoundInstanceIDs []string
	BoundBindings    map[string]brokerapi.BindDetails
	SyslogDrainURL   string
	RouteServiceURL  string
	BackupAgentURL   string
	VolumeMounts     []brokerapi.VolumeMount

	BindingFetchDetails domain.FetchBindingDetails
	UnbindingDetails    brokerapi.UnbindDetails

	InstanceLimit int

	ProvisionError            error
	BindError                 error
	UnbindError               error
	DeprovisionError          error
	LastOperationError        error
	LastBindingOperationError error
	UpdateError               error
	GetInstanceError          error
	GetBindingError           error

	BrokerCalled             bool
	LastOperationState       brokerapi.LastOperationState
	LastOperationDescription string

	AsyncAllowed bool

	ShouldReturnAsync     bool
	DashboardURL          string
	OperationDataToReturn string

	LastOperationInstanceID string
	LastOperationData       string

	ReceivedContext bool

	ServiceID string
	PlanID    string
}

type FakeAsyncServiceBroker struct {
	FakeServiceBroker
	ShouldProvisionAsync bool
}

type FakeAsyncOnlyServiceBroker struct {
	FakeServiceBroker
}

type FakeBrokerContextKeyType string

const (
	FakeBrokerContextDataKey  FakeBrokerContextKeyType = "test_context"
	FakeBrokerContextFailsKey FakeBrokerContextKeyType = "fails"
)

func (fakeBroker *FakeServiceBroker) Services(ctx context.Context) ([]brokerapi.Service, error) {
	fakeBroker.BrokerCalled = true

	if val, ok := ctx.Value(FakeBrokerContextDataKey).(bool); ok {
		fakeBroker.ReceivedContext = val
	}

	if val, ok := ctx.Value(FakeBrokerContextFailsKey).(bool); ok && val {
		return []brokerapi.Service{}, errors.New("something went wrong!")
	}

	return []brokerapi.Service{
		{
			ID:            "ff32ea32-cbe1-490d-a379-94b80b75d152",
			Name:          "cdr-services",
			Description:   "Cassandra service for application development and testing",
			Bindable:      true,
			PlanUpdatable: true,
			Plans: []brokerapi.ServicePlan{
				{
					ID:          "5bd12fff-a293-4f3a-a6c1-670defbee447",
					Name:        "system-admin",
					Description: "Manage instance or subscription pairs",
				},
				{
					ID:          "900ee451-e082-4b61-9157-03c095c7e884",
					Name:        "internal-component",
					Description: "Plan for communication between CDR internal components",
				},
			},
		},
	}, nil
}

func (fakeBroker *FakeServiceBroker) Provision(context context.Context, instanceID string, details brokerapi.ProvisionDetails, asyncAllowed bool) (brokerapi.ProvisionedServiceSpec, error) {
	fakeBroker.BrokerCalled = true

	if val, ok := context.Value(FakeBrokerContextDataKey).(bool); ok {
		fakeBroker.ReceivedContext = val
	}

	if fakeBroker.ProvisionError != nil {
		return brokerapi.ProvisionedServiceSpec{}, fakeBroker.ProvisionError
	}

	if len(fakeBroker.ProvisionedInstances) >= fakeBroker.InstanceLimit {
		return brokerapi.ProvisionedServiceSpec{}, brokerapi.ErrInstanceLimitMet
	}

	if _, ok := fakeBroker.ProvisionedInstances[instanceID]; !ok {
		fakeBroker.ProvisionedInstances[instanceID] = details
		return brokerapi.ProvisionedServiceSpec{DashboardURL: fakeBroker.DashboardURL}, nil
	}

	if reflect.DeepEqual(fakeBroker.ProvisionedInstances[instanceID], details) {
		return brokerapi.ProvisionedServiceSpec{AlreadyExists: true, DashboardURL: fakeBroker.DashboardURL}, nil
	}

	return brokerapi.ProvisionedServiceSpec{}, apiresponses.ErrInstanceAlreadyExists
}

func (fakeBroker *FakeAsyncServiceBroker) Provision(context context.Context, instanceID string, details brokerapi.ProvisionDetails, asyncAllowed bool) (brokerapi.ProvisionedServiceSpec, error) {
	fakeBroker.BrokerCalled = true

	if fakeBroker.ProvisionError != nil {
		return brokerapi.ProvisionedServiceSpec{}, fakeBroker.ProvisionError
	}

	if len(fakeBroker.ProvisionedInstances) >= fakeBroker.InstanceLimit {
		return brokerapi.ProvisionedServiceSpec{}, brokerapi.ErrInstanceLimitMet
	}

	if _, ok := fakeBroker.ProvisionedInstances[instanceID]; !ok {
		fakeBroker.ProvisionedInstances[instanceID] = details
		return brokerapi.ProvisionedServiceSpec{IsAsync: fakeBroker.ShouldProvisionAsync, DashboardURL: fakeBroker.DashboardURL, OperationData: fakeBroker.OperationDataToReturn}, nil
	}

	if reflect.DeepEqual(fakeBroker.ProvisionedInstances[instanceID], details) {
		return brokerapi.ProvisionedServiceSpec{IsAsync: fakeBroker.ShouldProvisionAsync, AlreadyExists: true, DashboardURL: fakeBroker.DashboardURL, OperationData: fakeBroker.OperationDataToReturn}, nil
	}

	return brokerapi.ProvisionedServiceSpec{}, apiresponses.ErrInstanceAlreadyExists
}

func (fakeBroker *FakeAsyncOnlyServiceBroker) Provision(context context.Context, instanceID string, details brokerapi.ProvisionDetails, asyncAllowed bool) (brokerapi.ProvisionedServiceSpec, error) {
	fakeBroker.BrokerCalled = true

	if fakeBroker.ProvisionError != nil {
		return brokerapi.ProvisionedServiceSpec{}, fakeBroker.ProvisionError
	}

	if len(fakeBroker.ProvisionedInstances) >= fakeBroker.InstanceLimit {
		return brokerapi.ProvisionedServiceSpec{}, brokerapi.ErrInstanceLimitMet
	}

	if _, ok := fakeBroker.ProvisionedInstances[instanceID]; ok {
		if reflect.DeepEqual(fakeBroker.ProvisionedInstances[instanceID], details) {
			return brokerapi.ProvisionedServiceSpec{IsAsync: asyncAllowed, AlreadyExists: true, DashboardURL: fakeBroker.DashboardURL}, nil
		}

		return brokerapi.ProvisionedServiceSpec{}, apiresponses.ErrInstanceAlreadyExists
	}

	if !asyncAllowed {
		return brokerapi.ProvisionedServiceSpec{}, brokerapi.ErrAsyncRequired
	}

	fakeBroker.ProvisionedInstances[instanceID] = details
	return brokerapi.ProvisionedServiceSpec{IsAsync: true, DashboardURL: fakeBroker.DashboardURL}, nil
}

func (fakeBroker *FakeServiceBroker) Update(context context.Context, instanceID string, details brokerapi.UpdateDetails, asyncAllowed bool) (brokerapi.UpdateServiceSpec, error) {
	fakeBroker.BrokerCalled = true

	if val, ok := context.Value(FakeBrokerContextDataKey).(bool); ok {
		fakeBroker.ReceivedContext = val
	}

	if fakeBroker.UpdateError != nil {
		return brokerapi.UpdateServiceSpec{}, fakeBroker.UpdateError
	}

	fakeBroker.UpdateDetails = details
	fakeBroker.UpdatedInstanceIDs = append(fakeBroker.UpdatedInstanceIDs, instanceID)
	fakeBroker.AsyncAllowed = asyncAllowed
	return brokerapi.UpdateServiceSpec{IsAsync: fakeBroker.ShouldReturnAsync, OperationData: fakeBroker.OperationDataToReturn, DashboardURL: fakeBroker.DashboardURL}, nil
}

func (fakeBroker *FakeServiceBroker) GetInstance(context context.Context, instanceID string, details domain.FetchInstanceDetails) (brokerapi.GetInstanceDetailsSpec, error) {
	fakeBroker.BrokerCalled = true

	if val, ok := context.Value(FakeBrokerContextDataKey).(bool); ok {
		fakeBroker.ReceivedContext = val
	}

	fakeBroker.InstanceFetchDetails = details
	fakeBroker.GetInstanceIDs = append(fakeBroker.GetInstanceIDs, instanceID)
	return brokerapi.GetInstanceDetailsSpec{
		ServiceID:    fakeBroker.ServiceID,
		PlanID:       fakeBroker.PlanID,
		DashboardURL: fakeBroker.DashboardURL,
		Parameters: map[string]any{
			"param1": "value1",
		},
	}, fakeBroker.GetInstanceError
}

func (fakeBroker *FakeServiceBroker) Deprovision(context context.Context, instanceID string, details brokerapi.DeprovisionDetails, asyncAllowed bool) (brokerapi.DeprovisionServiceSpec, error) {
	fakeBroker.BrokerCalled = true

	if val, ok := context.Value(FakeBrokerContextDataKey).(bool); ok {
		fakeBroker.ReceivedContext = val
	}

	if fakeBroker.DeprovisionError != nil {
		return brokerapi.DeprovisionServiceSpec{}, fakeBroker.DeprovisionError
	}

	fakeBroker.DeprovisionDetails = details
	fakeBroker.DeprovisionedInstanceIDs = append(fakeBroker.DeprovisionedInstanceIDs, instanceID)

	if _, ok := fakeBroker.ProvisionedInstances[instanceID]; ok {
		return brokerapi.DeprovisionServiceSpec{}, nil
	}
	return brokerapi.DeprovisionServiceSpec{IsAsync: false}, nil
}

func (fakeBroker *FakeAsyncOnlyServiceBroker) Deprovision(context context.Context, instanceID string, details brokerapi.DeprovisionDetails, asyncAllowed bool) (brokerapi.DeprovisionServiceSpec, error) {
	fakeBroker.BrokerCalled = true

	if fakeBroker.DeprovisionError != nil {
		return brokerapi.DeprovisionServiceSpec{IsAsync: true}, fakeBroker.DeprovisionError
	}

	if !asyncAllowed {
		return brokerapi.DeprovisionServiceSpec{IsAsync: true}, brokerapi.ErrAsyncRequired
	}

	fakeBroker.DeprovisionedInstanceIDs = append(fakeBroker.DeprovisionedInstanceIDs, instanceID)
	fakeBroker.DeprovisionDetails = details

	if _, ok := fakeBroker.ProvisionedInstances[instanceID]; ok {
		return brokerapi.DeprovisionServiceSpec{IsAsync: true, OperationData: fakeBroker.OperationDataToReturn}, nil
	}

	return brokerapi.DeprovisionServiceSpec{IsAsync: true, OperationData: fakeBroker.OperationDataToReturn}, brokerapi.ErrInstanceDoesNotExist
}

func (fakeBroker *FakeAsyncServiceBroker) Deprovision(context context.Context, instanceID string, details brokerapi.DeprovisionDetails, asyncAllowed bool) (brokerapi.DeprovisionServiceSpec, error) {
	fakeBroker.BrokerCalled = true

	if fakeBroker.DeprovisionError != nil {
		return brokerapi.DeprovisionServiceSpec{IsAsync: asyncAllowed}, fakeBroker.DeprovisionError
	}

	fakeBroker.DeprovisionedInstanceIDs = append(fakeBroker.DeprovisionedInstanceIDs, instanceID)
	fakeBroker.DeprovisionDetails = details

	if _, ok := fakeBroker.ProvisionedInstances[instanceID]; ok {
		return brokerapi.DeprovisionServiceSpec{IsAsync: asyncAllowed, OperationData: fakeBroker.OperationDataToReturn}, nil
	}

	return brokerapi.DeprovisionServiceSpec{OperationData: fakeBroker.OperationDataToReturn, IsAsync: asyncAllowed}, brokerapi.ErrInstanceDoesNotExist
}

func (fakeBroker *FakeServiceBroker) GetBinding(context context.Context, instanceID, bindingID string, details domain.FetchBindingDetails) (brokerapi.GetBindingSpec, error) {
	fakeBroker.BrokerCalled = true

	if val, ok := context.Value(FakeBrokerContextDataKey).(bool); ok {
		fakeBroker.ReceivedContext = val
	}

	fakeBroker.BindingFetchDetails = details
	return brokerapi.GetBindingSpec{
		Credentials: FakeCredentials{
			Host:     "127.0.0.1",
			Port:     3000,
			Username: "batman",
			Password: "robin",
		},
		SyslogDrainURL:  fakeBroker.SyslogDrainURL,
		RouteServiceURL: fakeBroker.RouteServiceURL,
		VolumeMounts:    fakeBroker.VolumeMounts,
	}, fakeBroker.GetBindingError
}

func (fakeBroker *FakeAsyncServiceBroker) Bind(context context.Context, instanceID, bindingID string, details brokerapi.BindDetails, asyncAllowed bool) (brokerapi.Binding, error) {
	fakeBroker.BrokerCalled = true

	if asyncAllowed {
		if _, ok := fakeBroker.BoundBindings[bindingID]; ok {
			return fakeBroker.FakeServiceBroker.Bind(context, instanceID, bindingID, details, true)
		}

		fakeBroker.BoundInstanceIDs = append(fakeBroker.BoundInstanceIDs, instanceID)
		fakeBroker.BoundBindings[bindingID] = details
		return brokerapi.Binding{
			IsAsync:       true,
			OperationData: "0xDEADBEEF",
		}, nil
	}

	return fakeBroker.FakeServiceBroker.Bind(context, instanceID, bindingID, details, false)
}

func (fakeBroker *FakeServiceBroker) Bind(context context.Context, instanceID, bindingID string, details brokerapi.BindDetails, asyncAllowed bool) (brokerapi.Binding, error) {
	fakeBroker.BrokerCalled = true

	if val, ok := context.Value(FakeBrokerContextDataKey).(bool); ok {
		fakeBroker.ReceivedContext = val
	}

	binding := brokerapi.Binding{
		Credentials: FakeCredentials{
			Host:     "127.0.0.1",
			Port:     3000,
			Username: "batman",
			Password: "robin",
		},
		SyslogDrainURL:  fakeBroker.SyslogDrainURL,
		RouteServiceURL: fakeBroker.RouteServiceURL,
		VolumeMounts:    fakeBroker.VolumeMounts,
	}

	if fakeBroker.BackupAgentURL != "" {
		binding = brokerapi.Binding{BackupAgentURL: fakeBroker.BackupAgentURL}
	}

	if _, ok := fakeBroker.BoundBindings[bindingID]; ok {
		if reflect.DeepEqual(fakeBroker.BoundBindings[bindingID], details) {
			binding.AlreadyExists = true
			return binding, nil
		}
	}

	if fakeBroker.BindError != nil {
		return brokerapi.Binding{}, fakeBroker.BindError
	}

	fakeBroker.BoundInstanceIDs = append(fakeBroker.BoundInstanceIDs, instanceID)
	fakeBroker.BoundBindings[bindingID] = details

	return binding, nil
}

func (fakeBroker *FakeServiceBroker) Unbind(context context.Context, instanceID, bindingID string, details brokerapi.UnbindDetails, asyncAllowed bool) (brokerapi.UnbindSpec, error) {
	fakeBroker.BrokerCalled = true

	if val, ok := context.Value(FakeBrokerContextDataKey).(bool); ok {
		fakeBroker.ReceivedContext = val
	}

	if fakeBroker.UnbindError != nil {
		return brokerapi.UnbindSpec{}, fakeBroker.UnbindError
	}

	fakeBroker.UnbindingDetails = details

	if _, ok := fakeBroker.ProvisionedInstances[instanceID]; ok {
		if _, ok := fakeBroker.BoundBindings[bindingID]; ok {
			return brokerapi.UnbindSpec{}, nil
		}
		return brokerapi.UnbindSpec{}, nil
	}

	return brokerapi.UnbindSpec{}, nil
}

func (fakeBroker *FakeServiceBroker) LastBindingOperation(context context.Context, instanceID, bindingID string, details brokerapi.PollDetails) (brokerapi.LastOperation, error) {

	if val, ok := context.Value(FakeBrokerContextDataKey).(bool); ok {
		fakeBroker.ReceivedContext = val
	}

	if fakeBroker.LastBindingOperationError != nil {
		return brokerapi.LastOperation{}, fakeBroker.LastBindingOperationError
	}

	return brokerapi.LastOperation{State: brokerapi.Succeeded, Description: fakeBroker.LastOperationDescription}, nil
}

func (fakeBroker *FakeServiceBroker) LastOperation(context context.Context, instanceID string, details brokerapi.PollDetails) (brokerapi.LastOperation, error) {
	fakeBroker.LastOperationInstanceID = instanceID
	fakeBroker.LastOperationData = details.OperationData

	if val, ok := context.Value(FakeBrokerContextDataKey).(bool); ok {
		fakeBroker.ReceivedContext = val
	}

	if fakeBroker.LastOperationError != nil {
		return brokerapi.LastOperation{}, fakeBroker.LastOperationError
	}

	return brokerapi.LastOperation{State: brokerapi.Succeeded, Description: fakeBroker.LastOperationDescription}, nil
}

type FakeCredentials struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func sliceContains(needle string, haystack []string) bool {
	return slices.Contains(haystack, needle)
}
