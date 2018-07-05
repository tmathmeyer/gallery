
import { GalleryImage, Gallery, SeedableRNG } from "./gallery";

export class GalleryTileImageLoader extends GalleryImage {

  constructor(url: string, private dim: [number, number], private title: string, private name: string) {
    super(url)
  }

  GetTypeName() : string {
    return "GalleryTileLoader";
  }

  GetDomElement() : HTMLElement {
    let _result = document.createElement('div');
    _result.classList.add('GalleryImage-' + this.GetTypeName());
    _result.classList.add('albumcover');
    _result.onclick = () => { location.href='/gallery/' + this.name };
    _result.style['background-image'] = 'url(' + this.imageUrl + ')';
    _result.style['background-size'] = 'cover';

    let _title = document.createElement('div');
    _title.classList.add('covertitle');
    _title.classList.add('albumletterpress');
    _title.innerText = this.title;

    _result.appendChild(_title)
    return _result;
  }

  GetAspectRatio() : number {
    return this.dim[0] / this.dim[1]
  }
}

export class GalleryDetailTileImageLoader extends GalleryImage {

  constructor(private gallery: string, private name: string, private type: number, private dim: [number, number]) {
    super('/img/' + gallery + '/' + name + '/VGA');
  }

  GetTypeName() : string {
    return "GalleryDetaleTileLoader";
  }

  GetDomElement() : HTMLElement {
    let _result = document.createElement('div');
    _result.classList.add('GalleryImage-' + this.GetTypeName());
    _result.classList.add('photo');
    _result.onclick = () => { this.LoadImage() };
    _result.style['background-image'] = 'url(' + this.imageUrl + ')';
    _result.style['background-size'] = 'cover';
    return _result;
  }

  private LoadImage() {
    let mobile = /android|webos|iphone|ipad|ipod|blackberry|iemobile|opera mini/i;
    if (mobile.test(navigator.userAgent.toLowerCase())) {
      window.open('/img/' + this.gallery + '/' + this.name + '/F', '_blank');
      return
    }

    if (this.type == 0) {
      window.open('/view/' + this.gallery + '/' + this.name, '_self');
      return
    }

    if (this.type == 1) {
      window.open('/pano/' + this.gallery + '/' + this.name, '_self');
      return
    }
  }

  GetAspectRatio() : number {
    return this.dim[0] / this.dim[1]
  }
}