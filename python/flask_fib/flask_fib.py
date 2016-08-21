from flask_opentracing import FlaskTracer
import flask
import lightstep.tracer
import opentracing
import requests

from ot_headers import get_outbound_headers

app = flask.Flask(__name__)

opentracing.tracer = lightstep.tracer.init_tracer(access_token="KYZXYZ")
flask_tracer = FlaskTracer(opentracing.tracer, True, app)


def fib_client(index):
    result_json = requests.post(
            "http://localhost:5000/add",
            json={"index": index},
            headers=get_outbound_headers(flask_tracer)).json()
    return result_json["val"]


@app.route("/add", methods=['POST'])
def add():
    """
    A recursive fibonnaci helper.
    """
    active_span = flask_tracer.get_span()
    req_json = flask.request.get_json()
    active_span.log_event("/add called", payload=req_json)
    index = req_json["index"]
    if index <= 1:
        active_span.log_event("base case (val=1)")
        return '{"val": 1}'
    else:
        val_l = fib_client(index - 1)
        val_r = fib_client(index - 2)
        active_span.log_event("return l+r", payload={"l": val_l, "r": val_r})
        return ('{"val": %d}\n' % (val_l + val_r,))


app.run(threaded=True)
