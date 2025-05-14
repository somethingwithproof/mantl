Release Process
===============

This document outlines the process for creating a new Mantl release.

Release Cycle
------------

Mantl follows a feature-based release cycle rather than a time-based cycle. New releases are created when significant features or improvements are ready.

Version Numbering
----------------

Mantl uses semantic versioning (MAJOR.MINOR.PATCH):

* MAJOR version for incompatible API changes
* MINOR version for backwards-compatible functionality
* PATCH version for backwards-compatible bug fixes

Pre-Release Planning
------------------

1. **Review open issues and PRs**:
   * Decide which issues and PRs should be included in the release
   * Prioritize critical bugs and security issues
   * Move non-critical issues to future milestones

2. **Create a release milestone**:
   * Create a milestone in GitHub
   * Assign relevant issues and PRs to the milestone
   * Set a target date for the release

3. **Feature freeze**:
   * Announce a feature freeze date
   * After the feature freeze, only bug fixes should be merged

Release Preparation
-----------------

1. **Documentation updates**:
   * Update documentation to reflect new features and changes
   * Update the upgrade guide if necessary
   * Review and refresh installation instructions

2. **Update version numbers**:
   * Update version in relevant files
   * Update changelog with all significant changes

3. **Testing**:
   * Run comprehensive tests on all supported platforms
   * Verify upgrade paths from previous versions
   * Test sample applications and workflows

4. **Create a release branch**:
   * Create a branch named ``release/X.Y.Z``
   * Make any release-specific changes in this branch

Release Procedure
---------------

1. **Final review**:
   * Review release notes and changelog
   * Verify all tests pass
   * Check documentation accuracy

2. **Create release candidate**:
   * Tag the release candidate as ``vX.Y.Z-rc1``
   * Build release artifacts if applicable
   * Announce the release candidate for testing

3. **Community testing**:
   * Allow time for community testing
   * Address any critical issues found
   * Create additional release candidates if necessary

4. **Create final release**:
   * Once testing is complete, tag the final release as ``vX.Y.Z``
   * Merge release branch to master
   * Push the tag to GitHub

5. **Publish release**:
   * Create a GitHub release with detailed release notes
   * Publish release artifacts to relevant platforms
   * Update documentation site with new version

Post-Release
-----------

1. **Announcement**:
   * Announce the release on the Mantl website
   * Send announcement to the Gitter channel
   * Share on social media and relevant community channels

2. **Clean up**:
   * Close the release milestone
   * Review and triage any issues that weren't addressed
   * Begin planning for the next release

3. **Retrospective**:
   * Review what went well and what could be improved
   * Document lessons learned for future releases

Release Checklist
---------------

Use this checklist for each release:

* [ ] Review and prioritize issues for release
* [ ] Create release milestone and assign issues
* [ ] Announce feature freeze
* [ ] Update documentation
* [ ] Update version numbers
* [ ] Update CHANGELOG.md
* [ ] Run comprehensive tests
* [ ] Create release branch
* [ ] Create release candidate(s)
* [ ] Community testing period
* [ ] Address critical issues
* [ ] Tag final release
* [ ] Merge to master
* [ ] Create GitHub release with notes
* [ ] Announce release
* [ ] Close milestone
* [ ] Conduct release retrospective

Emergency Releases
----------------

For critical security or bug fixes:

1. Fix the issue in a dedicated branch
2. Test thoroughly
3. Release as a patch version
4. Announce the security release with details about the fix

LTS Releases
-----------

Long-term support (LTS) releases receive bug fixes and security updates for an extended period:

* Designate specific releases as LTS
* Clearly communicate support timeframes
* Backport critical fixes to LTS branches