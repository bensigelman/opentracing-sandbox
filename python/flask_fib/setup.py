from setuptools import setup

setup(
    name='flask_fib',
    version='0.0.1.dev0',
    author='The OpenTracing Authors',
    author_email='info@opentracing.io',
    license='MIT',
    url='https://github.com/bensigelman/opentracing-sandbox',
    keywords=['opentracing'],
    classifiers=[
        'Development Status :: 3 - Alpha',
        'Intended Audience :: Developers',
        'License :: OSI Approved :: MIT License',
        'Programming Language :: Python :: 2.7',
        'Programming Language :: Python :: Implementation :: PyPy',
        'Topic :: Software Development :: Libraries :: Python Modules',
    ],
    packages=['lightstep', 'Flask-Opentracing'],
    include_package_data=True,
    zip_safe=False,
    platforms='any',
    install_requires=[
        'lightstep',
        'Flask-Opentracing',
        'opentracing>=1.1,<1.2',
        'requests',
    ],
)
