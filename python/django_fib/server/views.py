from django.shortcuts import render
from django.http import HttpResponse
from django.conf import settings

import opentracing
import requests

tracer = settings.OPENTRACING_TRACER

# Create your views here.

def server_index(request):
    return HttpResponse("Hello, world. You're at the server index.")

@tracer.trace()
def server_fib(request):
    span = tracer.get_span(request)
    if span is not None:
        with tracer._tracer.start_span("fib client", child_of=span.context) as rpc_span:
            requests.post(
                    "http://localhost:5000/add",
                    json={"index": 9},
                    headers=_ot_headers(rpc_span))
    return HttpResponse("A child span was created")


def _ot_headers(span):
    rval = {}
    tracer._tracer.inject(span.context, opentracing.Format.HTTP_HEADERS, rval)
    return rval
