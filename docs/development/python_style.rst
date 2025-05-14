Python Code Style Guide
=====================

This document describes the coding style for Python code in Mantl.

General Guidelines
-----------------

* Follow `PEP 8 <https://www.python.org/dev/peps/pep-0008/>`_ style guide
* Limit lines to 100 characters 
* Use 4 spaces for indentation (no tabs)
* Add docstrings to all public modules, functions, classes, and methods

Imports
-------

* Group imports in the following order:
  1. Standard library imports
  2. Related third party imports
  3. Local application/library specific imports
* Within each group, imports should be sorted alphabetically
* Always use absolute imports when possible

Example::

    # Standard library imports
    import json
    import os
    import sys
    
    # Third party imports
    import ansible
    import terraform
    
    # Local imports
    from mantl import utils

Documentation
------------

* Always include docstrings for public modules, classes, and functions
* Use `reStructuredText <http://docutils.sourceforge.net/rst.html>`_ format for docstrings
* Include parameter descriptions and return values

Example::

    def function_name(param1, param2):
        """Short description of the function.
        
        More detailed explanation of the function if needed.
        
        Args:
            param1: Description of param1
            param2: Description of param2
            
        Returns:
            Description of the return value
            
        Raises:
            ExceptionName: Description of when this exception is raised
        """
        # Function implementation

Naming Conventions
-----------------

* Use ``CamelCase`` for class names
* Use ``lowercase_with_underscores`` for function, method, and variable names
* Use ``UPPERCASE_WITH_UNDERSCORES`` for constants
* Prefix private attributes and methods with a single underscore (_)

Testing
------

* Write unit tests for all new code
* Use pytest for testing
* Maintain at least 80% code coverage for all new code

Code Linting
-----------

* Use flake8 or pylint to check code quality
* Fix any linting errors before submitting code

Example flake8 configuration::

    [flake8]
    max-line-length = 100
    exclude = .git,__pycache__,docs/source/conf.py,old,build,dist
    
Version Control
--------------

* Make small, focused commits
* Use descriptive commit messages
* Reference GitHub issues in commit messages when applicable

Additional Resources
------------------

* `Google Python Style Guide <https://google.github.io/styleguide/pyguide.html>`_
* `PEP 257 - Docstring Conventions <https://www.python.org/dev/peps/pep-0257/>`_