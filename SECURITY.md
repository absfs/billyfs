# Security Policy

## Supported Versions

The following versions of billyfs are currently supported with security updates:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |

## Reporting a Vulnerability

We take the security of billyfs seriously. If you discover a security vulnerability, please follow these steps:

### How to Report

1. **DO NOT** open a public GitHub issue for security vulnerabilities
2. Email the maintainers directly at: security@absfs.org (if available) or open a private security advisory on GitHub
3. Include the following information in your report:
   - Description of the vulnerability
   - Steps to reproduce the issue
   - Potential impact and attack scenarios
   - Any suggested fixes or mitigations (if available)

### What to Expect

- **Initial Response**: We will acknowledge receipt of your vulnerability report within 48 hours
- **Status Updates**: We will send you regular updates (at least every 5 business days) about our progress
- **Resolution Timeline**: We aim to release a fix within 30 days for critical vulnerabilities
- **Credit**: We will credit you in the security advisory (unless you prefer to remain anonymous)

### Disclosure Policy

- We follow a coordinated disclosure approach
- We will work with you to understand and resolve the issue before any public disclosure
- Once a fix is available, we will:
  1. Release a patched version
  2. Publish a security advisory on GitHub
  3. Update the CHANGELOG with security-related changes
  4. Credit the reporter (if they agree)

## Security Considerations

### Path Traversal

billyfs inherits path validation from the underlying absfs.SymlinkFileSystem implementation. When using billyfs:

- Always validate user input before passing paths to filesystem operations
- Be aware that the underlying filesystem may or may not prevent path traversal attacks
- Consider using the `Chroot` method to limit filesystem access to a specific directory tree
- Test your specific absfs implementation for path traversal vulnerabilities

Example of safe path handling:

```go
// Validate user input
userPath := getUserInput()
if strings.Contains(userPath, "..") {
    return errors.New("invalid path: contains directory traversal")
}

// Use Join instead of manual path construction
safePath := bfs.Join("/allowed/directory", filepath.Base(userPath))
```

### Temporary File Security

The `TempFile` method generates temporary files with random names. Consider these security practices:

- Temporary files are created with 0600 permissions (owner read/write only)
- File names include random components to prevent prediction attacks
- Clean up temporary files promptly after use
- Be aware of the underlying filesystem's temporary file handling

### Symlink Security

billyfs supports symbolic links through the underlying absfs.SymlinkFileSystem:

- Symlinks can potentially escape chroot boundaries depending on the underlying implementation
- Validate symlink targets before creation
- Use `Lstat` instead of `Stat` when you need to examine symlinks themselves
- Be cautious when following symlinks from untrusted sources

### Permission Handling

billyfs delegates permission handling to the underlying filesystem:

- Verify that your absfs implementation enforces permissions correctly
- Test permission boundaries in your specific deployment environment
- Remember that in-memory filesystems may not enforce real OS permissions

### Race Conditions

When using billyfs in concurrent environments:

- The adapter itself is safe for concurrent use if the underlying filesystem is thread-safe
- Use appropriate locking when performing check-then-act operations
- Be aware of TOCTOU (Time-of-check to time-of-use) vulnerabilities

Example of safe concurrent file creation:

```go
// Unsafe: check-then-act race condition
if _, err := bfs.Stat(path); os.IsNotExist(err) {
    // Race condition window here
    file, _ := bfs.Create(path)
}

// Safer: atomic operation
file, err := bfs.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
if err != nil {
    // File already exists or other error
}
```

## Security Testing

We maintain a suite of security-focused tests in `billyfs_security_test.go`. When contributing:

- Run security tests with: `go test -run Security`
- Consider race condition tests: `go test -race`
- Add security tests for new functionality
- Review tests for edge cases and attack vectors

## Dependencies

billyfs depends on:
- `github.com/absfs/absfs` - Abstract filesystem interface
- `github.com/absfs/basefs` - Base filesystem implementation
- `github.com/go-git/go-billy/v5` - Billy filesystem interface

We monitor these dependencies for security vulnerabilities using:
- GitHub Dependabot alerts
- Regular dependency updates
- Security advisories from upstream projects

### Enabling Dependabot

For repository maintainers, ensure Dependabot is enabled:

1. Go to repository Settings → Security & analysis
2. Enable "Dependabot alerts"
3. Enable "Dependabot security updates"

## Best Practices

When using billyfs in production:

1. **Input Validation**: Always validate and sanitize user-provided paths
2. **Least Privilege**: Use `Chroot` to restrict filesystem access to necessary directories
3. **Secure Defaults**: Use restrictive file permissions (0600 for files, 0700 for directories)
4. **Error Handling**: Don't expose internal filesystem paths in error messages to users
5. **Auditing**: Log filesystem operations in security-sensitive applications
6. **Testing**: Include security tests in your integration test suite
7. **Updates**: Keep billyfs and its dependencies up to date

## Known Security Considerations

### Filesystem Implementation Dependent

billyfs is an adapter and inherits the security characteristics of the underlying absfs.SymlinkFileSystem implementation. Security properties such as:

- Path validation
- Permission enforcement
- Symlink handling
- Concurrent access safety

All depend on the specific filesystem implementation you use. Always review the security documentation of your chosen absfs implementation.

### No Built-in Sandboxing

billyfs does not provide additional sandboxing beyond what the underlying filesystem offers. If you need strong isolation:

- Use OS-level containers or sandboxes
- Implement additional access control layers
- Consider using specialized filesystem implementations with built-in security features

## Contact

For security-related questions or concerns:
- Open a GitHub Discussion for general security questions
- Email security@absfs.org for private vulnerability reports (if available)
- Use GitHub's private security advisory feature

## Acknowledgments

We appreciate the security researchers and users who help keep billyfs secure. Contributors who responsibly disclose vulnerabilities will be acknowledged in our security advisories and release notes.
