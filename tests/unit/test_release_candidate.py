"""Local candidate validation rejects mismatched bytes and incomplete user assets."""

import hashlib
import io
import tarfile

import pytest

from ci import verify_release_candidate as candidate


def candidate_directory(tmp_path, monkeypatch):
    monkeypatch.setattr(candidate.release_paths, "WORKSPACE_ROOT", tmp_path)
    directory = tmp_path / "dist" / "candidate"
    directory.mkdir(parents=True)
    return directory


def test_candidate_checksums_require_exact_complete_inventory(tmp_path, monkeypatch):
    directory = candidate_directory(tmp_path, monkeypatch)
    version = "0.5.0-rc.1"
    lines = []
    for name in sorted(candidate.expected_artifacts(version)):
        data = name.encode()
        (directory / name).write_bytes(data)
        lines.append(f"{hashlib.sha256(data).hexdigest()}  {name}\n")
    manifest = directory / "cli-checksums.txt"
    manifest.write_text("".join(lines))
    assert len(candidate.verify_checksums(directory, version)) == 8
    manifest.write_text("".join(lines[:-1]))
    with pytest.raises(ValueError, match="incomplete"):
        candidate.verify_checksums(directory, version)
    manifest.write_text("".join(lines + [lines[0]]))
    with pytest.raises(ValueError, match="invalid"):
        candidate.verify_checksums(directory, version)
    manifest.write_text("".join(lines))
    (directory / sorted(candidate.expected_artifacts(version))[0]).write_bytes(b"changed")
    with pytest.raises(ValueError, match="mismatch"):
        candidate.verify_checksums(directory, version)


def test_candidate_archive_needs_working_quickstart_and_logo(tmp_path):
    path = tmp_path / "candidate.tar.gz"
    with tarfile.open(path, "w:gz") as archive:
        for name in candidate.ARCHIVE_FILES - {"docs/_static/mantl-mark.svg"}:
            data = b"fixture"
            info = tarfile.TarInfo(name)
            info.size = len(data)
            archive.addfile(info, io.BytesIO(data))
    with pytest.raises(ValueError, match="omits"):
        candidate.verify_archive(path)


@pytest.mark.parametrize(
    "version", ["0.5.0", "v0.5.0-rc.1", "00.5.0-rc.1", "0.5.0-rc.0", "../0.5.0-rc.1"]
)
def test_candidate_version_is_explicit_prerelease(version):
    with pytest.raises(ValueError, match="SemVer"):
        candidate.expected_artifacts(version)
