class SyncItem extends React.Component {

    constructor(props) {
        super(props);
        this.state = {
            syncStatus: "Unknown"
        };
    }

    componentDidMount() {
        // fetch state of component
        var url = window.location.href + "sync?name=" + this.props.name;
        fetch(url).then(this.updateSyncStatus);
    }

    render() {
        return (
            <li>sync state of {this.props.name}...TBD</li>
        )
    }

}

class SyncList extends React.Component {

    render() {
        return (
            <ul>
                <SyncItem name="Nokia"/>
                <SyncItem name="FatSecret"/>
            </ul>
        )
    }

   
}

ReactDOM.render(
      <SyncList/>,
      document.getElementById('root')
);
