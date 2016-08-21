import opentracing


def get_outbound_headers(flask_tracer):
    headers = {}
    active_span = flask_tracer.get_span()
    opentracing.tracer.inject(
        active_span.context,
        opentracing.Format.HTTP_HEADERS,
        headers)
    return headers


