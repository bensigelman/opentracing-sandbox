from django.conf.urls import url

from . import views

urlpatterns = [
	url(r'^$', views.server_index, name='index'),
	url(r'^fib', views.server_fib, name='fib'),
]
