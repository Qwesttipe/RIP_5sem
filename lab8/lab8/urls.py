from django.contrib import admin
from django.urls import path
from app.views import  ComputeAsyncView

urlpatterns = [
    path('admin/', admin.site.urls),
    path('api/compute/', ComputeAsyncView.as_view(), name='compute_async'),
]
