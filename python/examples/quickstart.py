import os

import opentracing
import some_opentracing_impl


def hello_world(tracer):
    """
    Start and finish a trivial Span that measures a "pass".
    """
    with tracer.start_span("hello_world") as span:
        pass


def hello_kids(tracer):
    """
    Create a parent Span ("mom") with ten children Spans.
    """
    # Start and finish (via `with`) a trivial span.
    with tracer.start_span("hello_kids") as mom:
        for x in range(10):
            with tracer.start_span("hello_mom", child_of=mom) as child_span:
                # Set a tag on each child span.
                child_span.set_tag("kid_id", str(x))


if __name__ == "__main__":
    # One-time setup
    try:
        opentracing.tracer = some_opentracing_impl.Tracer(...)

        # Call on the trivial examples above.
        hello_world(opentracing.tracer)
        hello_mom(opentracing.tracer)

    # One-time teardown
    finally:
        # NOTE: `flush()` is not part of the OpenTracing specification and is
        # thus not guaranteed to be implemented by all Tracers. Consult
        # documentation for your particular Tracer implementation to determine
        # whether a flush() (or sleep(), or similar) is necessary before
        # exiting the process.
        opentracing.tracer.flush()
