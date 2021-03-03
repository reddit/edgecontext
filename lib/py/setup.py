from setuptools import find_packages
from setuptools import setup

setup(
    name="reddit_edgecontext",
    description="reddit edge request context baggage",
    long_description=open("../../README.md").read(),
    long_description_content_type="text/markdown",
    url="https://github.com/reddit/edgecontext",
    project_urls={
        "Documentation": "https://reddit-edgecontext.readthedocs.io/",
    },
    author="reddit",
    license="BSD",
    use_scm_version={
        "root": "../../",
        "relative_to": __file__,
    },
    packages=find_packages(),
    python_requires=">=3.7",
    setup_requires=["setuptools_scm"],
    install_requires=[
        "baseplate>=1.5,<3.0",
        "pyjwt>=2.0.0,<3.0",
        "thrift-unofficial>=0.14,<1.0",
        "cryptography>=3.0,<4.0",
    ],
    package_data={"reddit_edgecontext": ["py.typed"]},
    zip_safe=True,
    classifiers=[
        "Development Status :: 5 - Production/Stable",
        "License :: OSI Approved :: BSD License",
        "Operating System :: POSIX :: Linux",
        "Programming Language :: Python",
        "Programming Language :: Python :: 3.7",
        "Programming Language :: Python :: 3.8",
        "Topic :: Software Development :: Libraries",
    ],
)
