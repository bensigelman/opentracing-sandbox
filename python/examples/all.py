import os

import opentracing
import lightstep.tracer


def hello_world(tracer):
    with tracer.start_span("hello_world") as span:
        for x in range(10):
            span.log_event("a log message", payload={"foo": "bar", "x": x})


def hello_mom(tracer):
    with tracer.start_span("hello_mom") as parent_span:
        for x in range(10):
            with tracer.start_span("hello_kid", child_of=parent_span) as child_span:
                child_span.set_tag("kid_id", str(x))


if __name__ == "__main__":
    # One-time setup
    try:
        opentracing.tracer = lightstep.tracer.init_tracer(access_token=os.getenv("LS_TOKEN"))

        # Examples
        hello_world(opentracing.tracer)
        hello_mom(opentracing.tracer)

    # One-time teardown
    finally:
        opentracing.tracer.flush()
