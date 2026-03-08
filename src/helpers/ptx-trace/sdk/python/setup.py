"""PTX-TRACE Python SDK setup."""

from setuptools import setup, find_packages

with open("README.md", "r", encoding="utf-8") as f:
    long_description = f.read()

setup(
    name="ptx-trace",
    version="1.0.0",
    author="Portunix Team",
    author_email="portunix@cassandragargoyle.cz",
    description="Python SDK for PTX-TRACE - Universal tracing system for software development",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/cassandragargoyle/portunix",
    packages=find_packages(),
    classifiers=[
        "Development Status :: 4 - Beta",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
        "Topic :: Software Development :: Debuggers",
        "Topic :: Software Development :: Testing",
        "Topic :: System :: Logging",
    ],
    python_requires=">=3.8",
    install_requires=[],
    extras_require={
        "dev": [
            "pytest>=7.0",
            "pytest-cov>=4.0",
            "mypy>=1.0",
            "black>=23.0",
            "isort>=5.0",
        ],
    },
    entry_points={
        "console_scripts": [],
    },
    keywords="tracing, debugging, etl, pipeline, logging, observability",
    project_urls={
        "Bug Reports": "https://github.com/cassandragargoyle/portunix/issues",
        "Source": "https://github.com/cassandragargoyle/portunix",
        "Documentation": "https://github.com/cassandragargoyle/portunix/tree/main/sdk/python",
    },
)
