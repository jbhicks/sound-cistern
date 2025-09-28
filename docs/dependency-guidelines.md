# Project Guidelines: External Dependencies and CMS Solutions

## CRITICAL DEPENDENCY RULES 🚨

**NEVER RECOMMEND NON-GO SOLUTIONS**

### What NOT to Recommend

❌ **Commercial/Corporate CMS Solutions:**
- Strapi (Node.js, commercial)
- Contentful (SaaS, commercial)
- Ghost (Node.js)
- WordPress (PHP)
- Drupal (PHP)
- Any proprietary CMS with corporate licensing

❌ **Non-Go Programming Languages:**
- Node.js/JavaScript solutions
- PHP solutions  
- Python solutions
- Ruby solutions
- Any solution not written in Go

❌ **External SaaS/Cloud Services:**
- Headless CMS services
- Third-party API services
- Commercial content management platforms

### What TO Recommend

✅ **Go-Based Open Source Solutions Only:**
- Ponzu CMS (github.com/ponzu-cms/ponzu)
- Hugo (static site generator - Go)
- Custom Go solutions using Buffalo
- Go-based headless CMS libraries
- Native Buffalo/Pop database solutions

### Verification Checklist

Before recommending ANY external dependency, verify:

1. **Language**: Is it written in Go? ✅ or ❌
2. **License**: Is it open source (MIT, Apache, BSD)? ✅ or ❌
3. **Architecture**: Does it integrate as a Go module? ✅ or ❌
4. **Commercial Status**: Is it free and non-commercial? ✅ or ❌
5. **Project Values**: Does it align with open-source, Go-first principles? ✅ or ❌

**ALL FIVE MUST BE ✅ TO PROCEED**

### Research Process

When user requests CMS or external service integration:

1. **First**: Search GitHub for "golang cms", "go cms", "buffalo cms"
2. **Verify**: Check repository language, license, and activity
3. **Validate**: Ensure it's a Go module that can be imported
4. **Test**: Verify it works with Buffalo framework
5. **Document**: Only proceed if all criteria are met

### Project Values

This Buffalo SaaS template is built on:
- **Go-first**: All dependencies must be Go-based
- **Open Source**: No commercial or proprietary solutions
- **Self-hosted**: No external SaaS dependencies
- **Simple**: Minimal external dependencies
- **Maintainable**: Solutions we can understand and modify

### Emergency Protocol

If you accidentally recommend a non-Go or commercial solution:

1. **STOP immediately**
2. **Remove all related code and files** 
3. **Update this documentation** with the lesson learned
4. **Research proper Go alternatives**
5. **Implement the correct solution**

## Implementation Notes

- Always check if a solution can be implemented natively in Buffalo first
- Prefer built-in Buffalo/Pop functionality over external dependencies
- When external Go modules are needed, prioritize well-maintained, popular libraries
- Document all dependencies and their purposes in the project README

---

**Remember**: This project is about Go, Buffalo, and open-source values. Stay true to these principles.
