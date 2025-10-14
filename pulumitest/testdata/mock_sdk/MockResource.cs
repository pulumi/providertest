using Pulumi;

namespace MockSdk
{
    public class MockResource : CustomResource
    {
        [Output("value")]
        public Output<string> Value { get; private set; } = null!;

        public MockResource(string name, string value, CustomResourceOptions? options = null)
            : base("mock:index:MockResource", name, new MockResourceArgs { Value = value }, options)
        {
        }
    }

    internal sealed class MockResourceArgs : ResourceArgs
    {
        [Input("value", required: true)]
        public Input<string> Value { get; set; } = null!;
    }
}
