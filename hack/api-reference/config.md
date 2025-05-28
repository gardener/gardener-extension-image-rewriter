<p>Packages:</p>
<ul>
<li>
<a href="#config.image-rewriter.extensions.gardener.cloud%2fv1alpha1">config.image-rewriter.extensions.gardener.cloud/v1alpha1</a>
</li>
</ul>
<h2 id="config.image-rewriter.extensions.gardener.cloud/v1alpha1">config.image-rewriter.extensions.gardener.cloud/v1alpha1</h2>
<p>
<p>Package v1alpha1 is a version of the API.</p>
</p>
Resource Types:
<ul></ul>
<h3 id="config.image-rewriter.extensions.gardener.cloud/v1alpha1.Configuration">Configuration
</h3>
<p>
<p>Configuration contains information about the registry service configuration.</p>
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
<code>overwrites</code></br>
<em>
<a href="#config.image-rewriter.extensions.gardener.cloud/v1alpha1.ImageOverwrite">
[]ImageOverwrite
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Overwrites configure the source and target images that should be replaced.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="config.image-rewriter.extensions.gardener.cloud/v1alpha1.Image">Image
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.image-rewriter.extensions.gardener.cloud/v1alpha1.ImageOverwrite">ImageOverwrite</a>, 
<a href="#config.image-rewriter.extensions.gardener.cloud/v1alpha1.TargetConfiguration">TargetConfiguration</a>)
</p>
<p>
<p>Image contains information about an image.</p>
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
<h3 id="config.image-rewriter.extensions.gardener.cloud/v1alpha1.ImageOverwrite">ImageOverwrite
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.image-rewriter.extensions.gardener.cloud/v1alpha1.Configuration">Configuration</a>)
</p>
<p>
<p>ImageOverwrite contains information about an image overwrite configuration.</p>
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
<a href="#config.image-rewriter.extensions.gardener.cloud/v1alpha1.Image">
Image
</a>
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
<a href="#config.image-rewriter.extensions.gardener.cloud/v1alpha1.TargetConfiguration">
[]TargetConfiguration
</a>
</em>
</td>
<td>
<p>Targets are the target images to replace the source with.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="config.image-rewriter.extensions.gardener.cloud/v1alpha1.TargetConfiguration">TargetConfiguration
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.image-rewriter.extensions.gardener.cloud/v1alpha1.ImageOverwrite">ImageOverwrite</a>)
</p>
<p>
<p>TargetConfiguration contains information about the target image configuration.</p>
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
<code>Image</code></br>
<em>
<a href="#config.image-rewriter.extensions.gardener.cloud/v1alpha1.Image">
Image
</a>
</em>
</td>
<td>
<p>
(Members of <code>Image</code> are embedded into this type.)
</p>
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
<code>region</code></br>
<em>
string
</em>
</td>
<td>
<p>Region is the region where the target image is located.</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <a href="https://github.com/ahmetb/gen-crd-api-reference-docs">gen-crd-api-reference-docs</a>
</em></p>
