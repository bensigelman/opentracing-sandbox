from django.conf.urls import url

from . import views

urlpatterns = [
    url(r'^$', views.client_index, name='index'),
    url(r'^fib', views.django_as_client, name='fib'),
]
