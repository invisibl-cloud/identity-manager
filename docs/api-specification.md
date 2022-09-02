<p>Packages:</p>
<ul>
<li>
<a href="#identity-manager.io%2fv1alpha1">identity-manager.io/v1alpha1</a>
</li>
</ul>
<h2 id="identity-manager.io/v1alpha1">identity-manager.io/v1alpha1</h2>
<p>
<p>Package v1alpha1 contains resources for identity-manager</p>
</p>
Resource Types:
<ul></ul>
<h3 id="identity-manager.io/v1alpha1.AwsRoleSpecPod">AwsRoleSpecPod
</h3>
<p>
<p>AwsRoleSpecPod defines the AWS&rsquo;s role spec pod</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>LabelSelector</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>
(Members of <code>LabelSelector</code> are embedded into this type.)
</p>
</td>
</tr>
<tr>
<td>
<code>namespace</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Namespace of the Pod</p>
</td>
</tr>
</tbody>
</table>
<h3 id="identity-manager.io/v1alpha1.AzureIdentity">AzureIdentity
</h3>
<p>
(<em>Appears on:</em>
<a href="#identity-manager.io/v1alpha1.WorkloadIdentityAzure">WorkloadIdentityAzure</a>)
</p>
<p>
<p>AzureIdentity is the definition of Azure&rsquo;s Identity</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>APIVersion of the identity</p>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Kind of the identity</p>
</td>
</tr>
<tr>
<td>
<code>metadata</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.Metadata">
Metadata
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Metadata of the identity</p>
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.AzureIdentitySpec">
AzureIdentitySpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Spec of the identity</p>
<br/>
<br/>
<table>
</table>
</td>
</tr>
</tbody>
</table>
<h3 id="identity-manager.io/v1alpha1.AzureIdentityBinding">AzureIdentityBinding
</h3>
<p>
(<em>Appears on:</em>
<a href="#identity-manager.io/v1alpha1.WorkloadIdentityAzure">WorkloadIdentityAzure</a>)
</p>
<p>
<p>AzureIdentityBinding is the definition of Azure Identity Binding</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>APIVersion of the IdentityBinding</p>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Kind of the IdentityBinding</p>
</td>
</tr>
<tr>
<td>
<code>metadata</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.Metadata">
Metadata
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Metadata of the IdentityBinding</p>
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.AzureIdentityBindingSpec">
AzureIdentityBindingSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Spec of the IdentityBinding</p>
<br/>
<br/>
<table>
</table>
</td>
</tr>
</tbody>
</table>
<h3 id="identity-manager.io/v1alpha1.AzureIdentityBindingSpec">AzureIdentityBindingSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#identity-manager.io/v1alpha1.AzureIdentityBinding">AzureIdentityBinding</a>)
</p>
<p>
<p>AzureIdentityBindingSpec defines the spec of the Identity Binding</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>selector</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Selector of the IdentityBinding</p>
</td>
</tr>
</tbody>
</table>
<h3 id="identity-manager.io/v1alpha1.AzureIdentitySpec">AzureIdentitySpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#identity-manager.io/v1alpha1.AzureIdentity">AzureIdentity</a>)
</p>
<p>
<p>AzureIdentitySpec defines the spec of the Identity</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>type</code></br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
<p>Type of the identity</p>
</td>
</tr>
</tbody>
</table>
<h3 id="identity-manager.io/v1alpha1.Condition">Condition
</h3>
<p>
(<em>Appears on:</em>
<a href="#identity-manager.io/v1alpha1.ConditionedStatus">ConditionedStatus</a>)
</p>
<p>
<p>A Condition that may apply to a resource.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>type</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.ConditionType">
ConditionType
</a>
</em>
</td>
<td>
<p>Type of this condition. At most one of each condition type may apply to
a resource at any point in time.</p>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#conditionstatus-v1-core">
Kubernetes core/v1.ConditionStatus
</a>
</em>
</td>
<td>
<p>Status of this condition; is it currently True, False, or Unknown?</p>
</td>
</tr>
<tr>
<td>
<code>lastTransitionTime</code></br>
<em>
<a href="https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#Time">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>LastTransitionTime is the last time this condition transitioned from one
status to another.</p>
</td>
</tr>
<tr>
<td>
<code>reason</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.ConditionReason">
ConditionReason
</a>
</em>
</td>
<td>
<p>A Reason for this condition&rsquo;s last transition from one status to another.</p>
</td>
</tr>
<tr>
<td>
<code>message</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>A Message containing details about this condition&rsquo;s last transition from
one status to another, if any.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="identity-manager.io/v1alpha1.ConditionReason">ConditionReason
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#identity-manager.io/v1alpha1.Condition">Condition</a>)
</p>
<p>
<p>A ConditionReason represents the reason a resource is in a condition.</p>
</p>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Available&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Creating&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Deleting&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;ReconcileError&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;ReconcileSuccess&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Unavailable&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="identity-manager.io/v1alpha1.ConditionType">ConditionType
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#identity-manager.io/v1alpha1.Condition">Condition</a>)
</p>
<p>
<p>A ConditionType represents a condition a resource could be in.</p>
</p>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Ready&#34;</p></td>
<td><p>TypeReady resources are believed to be ready to handle work.</p>
</td>
</tr><tr><td><p>&#34;Synced&#34;</p></td>
<td><p>TypeSynced resources are believed to be in sync with the
Kubernetes resources that manage their lifecycle.</p>
</td>
</tr></tbody>
</table>
<h3 id="identity-manager.io/v1alpha1.ConditionedStatus">ConditionedStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#identity-manager.io/v1alpha1.WorkloadIdentityStatus">WorkloadIdentityStatus</a>)
</p>
<p>
<p>A ConditionedStatus reflects the observed status of a resource. Only
one condition of each type may exist.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>conditions</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.Condition">
[]Condition
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Conditions of the resource.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="identity-manager.io/v1alpha1.Credentials">Credentials
</h3>
<p>
(<em>Appears on:</em>
<a href="#identity-manager.io/v1alpha1.WorkloadIdentitySpec">WorkloadIdentitySpec</a>)
</p>
<p>
<p>Credentials defines the credentials of the cloud provider</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>source</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.CredentialsSource">
CredentialsSource
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Source of the credentials</p>
</td>
</tr>
<tr>
<td>
<code>secretRef</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.SecretRef">
SecretRef
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SecretRef to fetch the credentials</p>
</td>
</tr>
<tr>
<td>
<code>properties</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Properties indicates extra properties of credentials</p>
</td>
</tr>
</tbody>
</table>
<h3 id="identity-manager.io/v1alpha1.CredentialsSource">CredentialsSource
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#identity-manager.io/v1alpha1.Credentials">Credentials</a>)
</p>
<p>
<p>A CredentialsSource is a source from which provider credentials may be
acquired.</p>
</p>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Secret&#34;</p></td>
<td><p>CredentialsSourceSecret indicates that a provider should acquire
credentials from a secret.</p>
</td>
</tr></tbody>
</table>
<h3 id="identity-manager.io/v1alpha1.ExternalResource">ExternalResource
</h3>
<p>
(<em>Appears on:</em>
<a href="#identity-manager.io/v1alpha1.WorkloadIdentityStatus">WorkloadIdentityStatus</a>)
</p>
<p>
<p>ExternalResource is the external resource&rsquo;s definition</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>id</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ID of the external resource</p>
</td>
</tr>
<tr>
<td>
<code>type</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Type of the external resource</p>
</td>
</tr>
</tbody>
</table>
<h3 id="identity-manager.io/v1alpha1.Metadata">Metadata
</h3>
<p>
(<em>Appears on:</em>
<a href="#identity-manager.io/v1alpha1.AzureIdentity">AzureIdentity</a>, 
<a href="#identity-manager.io/v1alpha1.AzureIdentityBinding">AzureIdentityBinding</a>)
</p>
<p>
<p>Metadata defines kubernetes resource&rsquo;s metadata</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Name of the Resource</p>
</td>
</tr>
<tr>
<td>
<code>namespace</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Namespace of the Resource</p>
</td>
</tr>
<tr>
<td>
<code>labels</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Labels of the Resource</p>
</td>
</tr>
<tr>
<td>
<code>annotations</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Annotations of the Resource</p>
</td>
</tr>
</tbody>
</table>
<h3 id="identity-manager.io/v1alpha1.Provider">Provider
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#identity-manager.io/v1alpha1.WorkloadIdentitySpec">WorkloadIdentitySpec</a>)
</p>
<p>
<p>Provider defines the cloud provider of the WorkloadIdentity</p>
</p>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;AWS&#34;</p></td>
<td><p>ProviderAWS is the AWS provider.</p>
</td>
</tr><tr><td><p>&#34;Azure&#34;</p></td>
<td><p>ProviderAzure is the Azure provider.</p>
</td>
</tr></tbody>
</table>
<h3 id="identity-manager.io/v1alpha1.Resource">Resource
</h3>
<p>
(<em>Appears on:</em>
<a href="#identity-manager.io/v1alpha1.WorkloadIdentityStatus">WorkloadIdentityStatus</a>)
</p>
<p>
<p>Resource is the definition of the kubernetes resource</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>APIVersion of the resource</p>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Kind of the resource</p>
</td>
</tr>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Name of the resource</p>
</td>
</tr>
<tr>
<td>
<code>namespace</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Namespace of the resource</p>
</td>
</tr>
</tbody>
</table>
<h3 id="identity-manager.io/v1alpha1.RoleAssignment">RoleAssignment
</h3>
<p>
(<em>Appears on:</em>
<a href="#identity-manager.io/v1alpha1.WorkloadIdentityAzure">WorkloadIdentityAzure</a>)
</p>
<p>
<p>RoleAssignment defines the role assignment</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>role</code></br>
<em>
string
</em>
</td>
<td>
<p>Role of the role assignment</p>
</td>
</tr>
<tr>
<td>
<code>scope</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Scope of the role assignment</p>
</td>
</tr>
</tbody>
</table>
<h3 id="identity-manager.io/v1alpha1.RoleDefinition">RoleDefinition
</h3>
<p>
<p>RoleDefinition is the definition for a Role</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>id</code></br>
<em>
string
</em>
</td>
<td>
<p>ID of the role definition (this will be used to generate internal UUID for role)</p>
</td>
</tr>
<tr>
<td>
<code>roleName</code></br>
<em>
string
</em>
</td>
<td>
<p>RoleName of the role definition</p>
</td>
</tr>
<tr>
<td>
<code>roleType</code></br>
<em>
string
</em>
</td>
<td>
<p>RoleType of the role definition</p>
</td>
</tr>
<tr>
<td>
<code>description</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Description of the role definition</p>
</td>
</tr>
<tr>
<td>
<code>assignableScopes</code></br>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>AssignableScopes is a list of assignable scopes</p>
</td>
</tr>
<tr>
<td>
<code>permissions</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.RolePermission">
[]RolePermission
</a>
</em>
</td>
<td>
<p>Permissions of the role definition</p>
</td>
</tr>
</tbody>
</table>
<h3 id="identity-manager.io/v1alpha1.RolePermission">RolePermission
</h3>
<p>
(<em>Appears on:</em>
<a href="#identity-manager.io/v1alpha1.RoleDefinition">RoleDefinition</a>)
</p>
<p>
<p>RolePermission defines the permissions of a Role</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>actions</code></br>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Actions is a list of actions</p>
</td>
</tr>
<tr>
<td>
<code>notActions</code></br>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>NotActions is a list of not actions</p>
</td>
</tr>
<tr>
<td>
<code>dataActions</code></br>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>DataActions is a list of data actions</p>
</td>
</tr>
<tr>
<td>
<code>notDataActions</code></br>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>NotDataActions is a list of not data actions</p>
</td>
</tr>
</tbody>
</table>
<h3 id="identity-manager.io/v1alpha1.SecretRef">SecretRef
</h3>
<p>
(<em>Appears on:</em>
<a href="#identity-manager.io/v1alpha1.Credentials">Credentials</a>)
</p>
<p>
<p>SecretRef defines the reference to the secret</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>namespace</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Namespace of the secret.</p>
</td>
</tr>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
<p>Name of the secret.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="identity-manager.io/v1alpha1.ServiceAccount">ServiceAccount
</h3>
<p>
<p>ServiceAccount defines the service account&rsquo;s metadata</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>action</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.ServiceAccountAction">
ServiceAccountAction
</a>
</em>
</td>
<td>
<p>Action to be perform on ServiceAccount</p>
</td>
</tr>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Name of the ServiceAccount</p>
</td>
</tr>
<tr>
<td>
<code>namespace</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Namespace of the ServiceAccount</p>
</td>
</tr>
<tr>
<td>
<code>Annotations</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Annotations to be added on ServiceAccount</p>
</td>
</tr>
</tbody>
</table>
<h3 id="identity-manager.io/v1alpha1.ServiceAccountAction">ServiceAccountAction
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#identity-manager.io/v1alpha1.ServiceAccount">ServiceAccount</a>)
</p>
<p>
<p>A ServiceAccountAction indicates action to be perform on ServiceAccount</p>
</p>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Create&#34;</p></td>
<td><p>ServiceAccountActionCreate indicates create service account</p>
</td>
</tr><tr><td><p>&#34;&#34;</p></td>
<td><p>ServiceAccountActionDefault indicates no action</p>
</td>
</tr><tr><td><p>&#34;Update&#34;</p></td>
<td><p>ServiceAccountActionUpdate indicates updating service account</p>
</td>
</tr></tbody>
</table>
<h3 id="identity-manager.io/v1alpha1.WorkloadIdentity">WorkloadIdentity
</h3>
<p>
<p>WorkloadIdentity is the Schema for the workloadidentities API</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.WorkloadIdentitySpec">
WorkloadIdentitySpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Name of the WorkloadIdentity</p>
</td>
</tr>
<tr>
<td>
<code>description</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Desc of the WorkloadIdentity</p>
</td>
</tr>
<tr>
<td>
<code>credentials</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.Credentials">
Credentials
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Credentials to manage the WorkloadIdentity</p>
</td>
</tr>
<tr>
<td>
<code>provider</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.Provider">
Provider
</a>
</em>
</td>
<td>
<p>Provider of the WorkloadIdentity</p>
</td>
</tr>
<tr>
<td>
<code>aws</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.WorkloadIdentityAWS">
WorkloadIdentityAWS
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>AWS WorkloadIdentity</p>
</td>
</tr>
<tr>
<td>
<code>azure</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.WorkloadIdentityAzure">
WorkloadIdentityAzure
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Azure WorkloadIdentity</p>
</td>
</tr>
<tr>
<td>
<code>writeToSecretRef</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.WriteToSecretRef">
WriteToSecretRef
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>WriteToSecretRef is a reference to a secret</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.WorkloadIdentityStatus">
WorkloadIdentityStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="identity-manager.io/v1alpha1.WorkloadIdentityAWS">WorkloadIdentityAWS
</h3>
<p>
(<em>Appears on:</em>
<a href="#identity-manager.io/v1alpha1.WorkloadIdentitySpec">WorkloadIdentitySpec</a>)
</p>
<p>
<p>WorkloadIdentityAWS defines the spec for AWS Provider</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>path</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Path of the Role</p>
</td>
</tr>
<tr>
<td>
<code>maxSessionDuration</code></br>
<em>
int64
</em>
</td>
<td>
<em>(Optional)</em>
<p>MaxSessionDuration of the Role</p>
</td>
</tr>
<tr>
<td>
<code>assumeRolePolicy</code></br>
<em>
string
</em>
</td>
<td>
<p>AssumeRolePolicy of the Role</p>
</td>
</tr>
<tr>
<td>
<code>inlinePolicies</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>InlinePolicies of the Role</p>
</td>
</tr>
<tr>
<td>
<code>policies</code></br>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Policies of the Role</p>
</td>
</tr>
<tr>
<td>
<code>serviceAccounts</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.*github.com/invisibl-cloud/identity-manager/api/v1alpha1.ServiceAccount">
[]*github.com/invisibl-cloud/identity-manager/api/v1alpha1.ServiceAccount
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ServiceAccounts to be managed</p>
</td>
</tr>
<tr>
<td>
<code>pods</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.*github.com/invisibl-cloud/identity-manager/api/v1alpha1.AwsRoleSpecPod">
[]*github.com/invisibl-cloud/identity-manager/api/v1alpha1.AwsRoleSpecPod
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Pods to be managed</p>
</td>
</tr>
</tbody>
</table>
<h3 id="identity-manager.io/v1alpha1.WorkloadIdentityAzure">WorkloadIdentityAzure
</h3>
<p>
(<em>Appears on:</em>
<a href="#identity-manager.io/v1alpha1.WorkloadIdentitySpec">WorkloadIdentitySpec</a>)
</p>
<p>
<p>WorkloadIdentityAzure is the Provider spec for ProviderAzure</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>roleDefinitions</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.*github.com/invisibl-cloud/identity-manager/api/v1alpha1.RoleDefinition">
[]*github.com/invisibl-cloud/identity-manager/api/v1alpha1.RoleDefinition
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RoleDefinitions is a list of role definitions</p>
</td>
</tr>
<tr>
<td>
<code>roleAssignments</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.RoleAssignment">
map[string]github.com/invisibl-cloud/identity-manager/api/v1alpha1.RoleAssignment
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RoleAssignments of the WorkloadIdentity</p>
</td>
</tr>
<tr>
<td>
<code>identity</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.AzureIdentity">
AzureIdentity
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Identity of the WorkloadIdentity</p>
</td>
</tr>
<tr>
<td>
<code>identityBinding</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.AzureIdentityBinding">
AzureIdentityBinding
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>IdentityBinding of the WorkloadIdentity</p>
</td>
</tr>
</tbody>
</table>
<h3 id="identity-manager.io/v1alpha1.WorkloadIdentitySpec">WorkloadIdentitySpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#identity-manager.io/v1alpha1.WorkloadIdentity">WorkloadIdentity</a>)
</p>
<p>
<p>WorkloadIdentitySpec defines the desired state of WorkloadIdentity</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Name of the WorkloadIdentity</p>
</td>
</tr>
<tr>
<td>
<code>description</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Desc of the WorkloadIdentity</p>
</td>
</tr>
<tr>
<td>
<code>credentials</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.Credentials">
Credentials
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Credentials to manage the WorkloadIdentity</p>
</td>
</tr>
<tr>
<td>
<code>provider</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.Provider">
Provider
</a>
</em>
</td>
<td>
<p>Provider of the WorkloadIdentity</p>
</td>
</tr>
<tr>
<td>
<code>aws</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.WorkloadIdentityAWS">
WorkloadIdentityAWS
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>AWS WorkloadIdentity</p>
</td>
</tr>
<tr>
<td>
<code>azure</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.WorkloadIdentityAzure">
WorkloadIdentityAzure
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Azure WorkloadIdentity</p>
</td>
</tr>
<tr>
<td>
<code>writeToSecretRef</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.WriteToSecretRef">
WriteToSecretRef
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>WriteToSecretRef is a reference to a secret</p>
</td>
</tr>
</tbody>
</table>
<h3 id="identity-manager.io/v1alpha1.WorkloadIdentityStatus">WorkloadIdentityStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#identity-manager.io/v1alpha1.WorkloadIdentity">WorkloadIdentity</a>)
</p>
<p>
<p>WorkloadIdentityStatus defines the observed state of WorkloadIdentity</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>ConditionedStatus</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.ConditionedStatus">
ConditionedStatus
</a>
</em>
</td>
<td>
<p>
(Members of <code>ConditionedStatus</code> are embedded into this type.)
</p>
</td>
</tr>
<tr>
<td>
<code>id</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ID of the Identity</p>
</td>
</tr>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Name of the Identity</p>
</td>
</tr>
<tr>
<td>
<code>resources</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.Resource">
[]Resource
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Resources managed by the Identity</p>
</td>
</tr>
<tr>
<td>
<code>externalResources</code></br>
<em>
<a href="#identity-manager.io/v1alpha1.ExternalResource">
[]ExternalResource
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>External Resources managed bu the Identity</p>
</td>
</tr>
</tbody>
</table>
<h3 id="identity-manager.io/v1alpha1.WriteToSecretRef">WriteToSecretRef
</h3>
<p>
(<em>Appears on:</em>
<a href="#identity-manager.io/v1alpha1.WorkloadIdentitySpec">WorkloadIdentitySpec</a>)
</p>
<p>
<p>WriteToSecretRef is a reference to a secret</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
<p>Name of the secret</p>
</td>
</tr>
<tr>
<td>
<code>namespace</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Namespace of the secret</p>
</td>
</tr>
<tr>
<td>
<code>templateData</code></br>
<em>
map[string]string
</em>
</td>
<td>
<p>TemplateData is a template for the data to be written to the secret</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>.
</em></p>
