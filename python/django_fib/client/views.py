from django.conf import settings
from django.http import HttpResponse
from django.shortcuts import render

import opentracing
import urllib2

tracer = settings.OPENTRACING_TRACER

# Create your views here.

def client_index(request):
    return HttpResponse("Client index page")

@tracer.trace()
def client_fib(request):
    url = "http://localhost:8000/server/fib"
    new_request = urllib2.Request(url)
    current_span = tracer.get_span(request)
    _inject_as_headers(current_span, new_request)
    try:
        response = urllib2.urlopen(new_request)
        return HttpResponse(
                "Sent a fib request to django server; response=%s" %
                str(response))
    except urllib2.URLError as e:  
        return HttpResponse("Error: " + str(e))


def _inject_as_headers(span, request):
    text_carrier = {}
    tracer._tracer.inject(span.context, opentracing.Format.HTTP_HEADERS, text_carrier)
    for k, v in text_carrier.iteritems():
        request.add_header(k,v)

