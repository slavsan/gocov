package internal

const Script = `const node = JSON.parse(document.querySelector('.tree-data').textContent)
const table = document.querySelector('.table')
const indicator = document.querySelector('.indicator')
const _ = null

let currentHash = location.hash.replace('#', '')

renderTable(currentHash)

function renderBreadcrumbs(currentHash) {
    const breadcrumbs = document.querySelector('.breadcrumbs')
    breadcrumbs.innerHTML = ''
    const allFiles = document.createElement('a')
    allFiles.textContent = 'All Files'
    allFiles.href = '#'
    breadcrumbs.appendChild(allFiles)
    const parts = currentHash.split('/')
    let links = []
    if (parts.length === 1 && (parts[0] === '#' || parts[0] === '')) {
        return
    }
    const len = parts.length
    parts.forEach((p, i) => {
        const divider = document.createElement('span')
        divider.textContent = '/'
        breadcrumbs.appendChild(divider)
        if (i === len-1) {
            const span = document.createElement('span')
            span.textContent = p
            breadcrumbs.appendChild(span)
            return
        }
        const a = document.createElement('a')
        links.push(p)
        a.href = '#' + links.join('/')
        a.textContent = p
        breadcrumbs.appendChild(a)
    })
}

function renderSelected(node, currentHash) {
    if (node.path === currentHash) {
        const stats = document.querySelector('.stats')
        stats.textContent = node.covered + '/' + node.all + ' (' + node.percent + '%)'

        renderIndicator(node.percent)

        if (node.children) {
            node.children.forEach(c => {
                renderRow(c, currentHash)
            })
        }
        return
    }

    if (node.children) {
        node.children.forEach(c => {
            renderSelected(c, currentHash)
        })
    }
}

function renderIndicator(percent) {
    indicator.classList.remove('ok')
    indicator.classList.remove('warn')
    indicator.classList.remove('error')
    if (percent >= 80) {
        indicator.classList.add('ok')
    } else if (percent >= 50) {
        indicator.classList.add('warn')
    } else {
        indicator.classList.add('error')
    }
}

function renderNode(node, currentHash) {
    renderRow(node, currentHash)
    if (!node.children) {
        return
    }
    node.children.forEach(c => {
        renderNode(c, currentHash)
    })
}

function renderTable(currentHash) {
    document.querySelectorAll('.source').forEach(e => {
        e.classList.remove('visible')
    })
    if (currentHash) {
        const targetFileSource = document.getElementById(currentHash)
        if (targetFileSource) {
            targetFileSource.classList.add('visible')
        }
    }

    const table = document.querySelector('.table')
    while (table.firstChild) {
        table.removeChild(table.lastChild)
    }

    renderBreadcrumbs(currentHash)

    if (!currentHash) {
        const stats = document.querySelector('.stats')
        stats.textContent = node.covered + '/' + node.all + ' (' + node.percent + '%)'
        renderIndicator(node.percent)
        renderNode(node, currentHash)
    } else {
        renderSelected(node, currentHash)
    }
}

window.addEventListener('hashchange', function (e) {
    const parts = e.newURL.split('#')
    let currentHash = ''
    if (parts.length === 2) {
        currentHash = parts[1]
    }
    renderTable(currentHash)
})

function renderRow(node, currentHash) {
    e('tr', _, { onInit: (tr) => {
        if (node.percent >= 80) {
            tr.classList.add('ok')
        } else if (node.percent >= 50) {
            tr.classList.add('warn')
        } else {
            tr.classList.add('error')
        }
        table.appendChild(tr)
    }}, [
        e('td', _, { onInit: (td) => {
            if (!currentHash) {
                td.style.paddingLeft = (5 + (20 * node.level)) + "px"
            }
        }}, [
            e('a', _, { onInit: (a) => {
                a.textContent = node.name
                a.href = '#' + node.path
            }})
        ]),
        e('td', _, { onInit: (td) => {
            td.textContent = node.covered + "/" + node.all
        }}),
        e('td', _, { onInit: (td) => {
            td.textContent = node.percent + '%'
        }}),
        e('td', _, _, [
            e('div', _, { onInit: (div) => {
                div.classList.add('progress')
            }}, [
                e('div', _, { onInit: (div) => {
                    div.style = 'width: '+node.percent+'%;'
                }})
            ])
        ]),
    ])
}

function e(elType, attributes, options, children) {
    const element = document.createElement(elType)
    if (attributes) {
        for (const [name, value] of Object.entries(attributes)) {
            element.setAttribute(name, value)
        }
    }
    if (options) {
        if (options.onInit) options.onInit(element)
    }
    if (children && children.length) {
        children.forEach(childElement=> element.appendChild(childElement))
    }
    if (options) {
        if (options.parent) {
            options.parent.appendChild(element)
        }
    }
    return element
}
`
