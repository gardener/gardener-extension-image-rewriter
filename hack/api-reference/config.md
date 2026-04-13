<p>Packages:</p>
<ul>
<li>
<a href="#config.image-rewriter.extensions.gardener.cloud%2fv1alpha1">config.image-rewriter.extensions.gardener.cloud/v1alpha1</a>
</li>
</ul>

<h2 id="config.image-rewriter.extensions.gardener.cloud/v1alpha1">config.image-rewriter.extensions.gardener.cloud/v1alpha1</h2>
<p>

</p>

<h3 id="configuration">Configuration
</h3>


<p>
Configuration contains information about the registry service configuration.
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
<code>containerd</code></br>
<em>
<a href="#containerdconfiguration">ContainerdConfiguration</a> array
</em>
</td>
<td>
<em>(Optional)</em>
<p>ContainerdConfiguration contains the containerd configuration for the image rewriter.</p>
</td>
</tr>
<tr>
<td>
<code>overwrites</code></br>
<em>
<a href="#imageoverwrite">ImageOverwrite</a> array
</em>
</td>
<td>
<em>(Optional)</em>
<p>Overwrites configure the source and target images that should be replaced.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="containerdconfiguration">ContainerdConfiguration
</h3>


<p>
(<em>Appears on:</em><a href="#configuration">Configuration</a>)
</p>

<p>
ContainerdConfiguration contains information about a containerd upstream configuration.
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
<code>upstream</code></br>
<em>
string
</em>
</td>
<td>
<p>Upstream is the upstream name of the registry.</p>
</td>
</tr>
<tr>
<td>
<code>server</code></br>
<em>
string
</em>
</td>
<td>
<p>Server is the URL of the upstream registry.</p>
</td>
</tr>
<tr>
<td>
<code>hosts</code></br>
<em>
<a href="#containerdhostconfig">ContainerdHostConfig</a> array
</em>
</td>
<td>
<p>Hosts are the containerd hosts separated by provider and regions.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="containerdhostconfig">ContainerdHostConfig
</h3>


<p>
(<em>Appears on:</em><a href="#containerdconfiguration">ContainerdConfiguration</a>)
</p>

<p>
ContainerdHostConfig contains information about a containerd host configuration.
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
<code>url</code></br>
<em>
string
</em>
</td>
<td>
<p></p>
</td>
</tr>
<tr>
<td>
<code>provider</code></br>
<em>
string
</em>
</td>
<td>
<p>Provider is the name of the provider for which this target is applicable.</p>
</td>
</tr>
<tr>
<td>
<code>regions</code></br>
<em>
string array
</em>
</td>
<td>
<em>(Optional)</em>
<p>Regions are the regions where the target image is located. If not specified, any shoot region will match this host config.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="image">Image
</h3>


<p>
(<em>Appears on:</em><a href="#imageoverwrite">ImageOverwrite</a>, <a href="#targetconfiguration">TargetConfiguration</a>)
</p>

<p>
Image contains information about an image.
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
<code>image</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Image is the target image string to relace the source with.</p>
</td>
</tr>
<tr>
<td>
<code>prefix</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Prefix is the prefix of the target image to relace the source with.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="imageoverwrite">ImageOverwrite
</h3>


<p>
(<em>Appears on:</em><a href="#configuration">Configuration</a>)
</p>

<p>
ImageOverwrite contains information about an image overwrite configuration.
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
<a href="#image">Image</a>
</em>
</td>
<td>
<p>Source is the source image string to be replaced.</p>
</td>
</tr>
<tr>
<td>
<code>targets</code></br>
<em>
<a href="#targetconfiguration">TargetConfiguration</a> array
</em>
</td>
<td>
<p>Targets are the target images to replace the source with.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="targetconfiguration">TargetConfiguration
</h3>


<p>
(<em>Appears on:</em><a href="#imageoverwrite">ImageOverwrite</a>)
</p>

<p>
TargetConfiguration contains information about the target image configuration.
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
<code>image</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Image is the target image string to relace the source with.</p>
</td>
</tr>
<tr>
<td>
<code>prefix</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Prefix is the prefix of the target image to relace the source with.</p>
</td>
</tr>
<tr>
<td>
<code>provider</code></br>
<em>
string
</em>
</td>
<td>
<p>Provider is the name of the provider for which this target is applicable.</p>
</td>
</tr>
<tr>
<td>
<code>regions</code></br>
<em>
string array
</em>
</td>
<td>
<em>(Optional)</em>
<p>Regions are the regions where the target image is located. If not specified, any shoot region will match this target config.</p>
</td>
</tr>

</tbody>
</table>


